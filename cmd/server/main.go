// Copyright 2019 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	app "github.com/moov-io/accounts"
	"github.com/moov-io/base/admin"
	moovhttp "github.com/moov-io/base/http"
	"github.com/moov-io/base/http/bind"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
	"github.com/mattn/go-sqlite3"
)

var (
	httpAddr  = flag.String("http.addr", bind.HTTP("accounts"), "HTTP listen address")
	adminAddr = flag.String("admin.addr", bind.Admin("accounts"), "Admin HTTP listen address")

	flagLogFormat = flag.String("log.format", "", "Format for log lines (Options: json, plain")
)

func main() {
	flag.Parse()

	fmt.Printf("http: %q\n", bind.HTTP("accounts"))
	fmt.Printf("admin: %q\n", bind.Admin("accounts"))

	fmt.Printf("http: %q\n", bind.HTTP("gl"))
	fmt.Printf("admin: %q\n", bind.Admin("gl"))

	var logger log.Logger
	if strings.ToLower(*flagLogFormat) == "json" {
		logger = log.NewJSONLogger(os.Stderr)
	} else {
		logger = log.NewLogfmtLogger(os.Stderr)
	}
	logger = log.With(logger, "ts", log.DefaultTimestampUTC)
	logger = log.With(logger, "caller", log.DefaultCaller)

	logger.Log("startup", fmt.Sprintf("Starting moov/accounts server version %s", app.Version))

	// Check for default routing number
	if defaultRoutingNumber == "" { // accounts.go
		logger.Log("main", "No default routing number specified, please set DEFAULT_ROUTING_NUMBER")
		os.Exit(1)
	}

	// Channel for errors
	errs := make(chan error)

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errs <- fmt.Errorf("%s", <-c)
	}()

	// Setup SQLite database
	if sqliteVersion, _, _ := sqlite3.Version(); sqliteVersion != "" {
		logger.Log("main", fmt.Sprintf("sqlite version %s", sqliteVersion))
	}
	db, err := createSqliteConnection(logger, getSqlitePath())
	if err != nil {
		logger.Log("main", err)
		os.Exit(1)
	}
	if err := migrate(logger, db); err != nil {
		logger.Log("main", err)
		os.Exit(1)
	}
	defer func() {
		if err := db.Close(); err != nil {
			logger.Log("main", err)
		}
	}()

	// Start Admin server (with Prometheus metrics)
	adminServer := admin.NewServer(*adminAddr)
	go func() {
		logger.Log("admin", fmt.Sprintf("listening on %s", adminServer.BindAddr()))
		if err := adminServer.Listen(); err != nil {
			err = fmt.Errorf("problem starting admin http: %v", err)
			logger.Log("admin", err)
			errs <- err
		}
	}()
	defer adminServer.Shutdown()

	// Setup Account storage
	accountStorageType := os.Getenv("ACCOUNT_STORAGE_TYPE")
	if accountStorageType == "" {
		accountStorageType = "sqlite"
	}
	accountRepo, err := initAccountStorage(logger, accountStorageType)
	if err != nil {
		panic(fmt.Sprintf("account storage: %v", err))
	}
	defer accountRepo.Close()
	logger.Log("main", fmt.Sprintf("using %T for account storage", accountRepo))
	adminServer.AddLivenessCheck(fmt.Sprintf("%s-accounts", accountStorageType), accountRepo.Ping)

	// Setup Transaction storage
	transactionStorageType := os.Getenv("TRANSACTION_STORAGE_TYPE")
	if transactionStorageType == "" {
		transactionStorageType = "sqlite"
	}
	transactionRepo, err := initTransactionStorage(logger, transactionStorageType)
	if err != nil {
		panic(fmt.Sprintf("transaction storage: %v", err))
	}
	defer transactionRepo.Close()
	logger.Log("main", fmt.Sprintf("using %T for transaction storage", transactionRepo))
	adminServer.AddLivenessCheck(fmt.Sprintf("%s-transactions", transactionStorageType), transactionRepo.Ping)

	// Setup business HTTP routes
	router := mux.NewRouter()
	moovhttp.AddCORSHandler(router)
	addPingRoute(logger, router)
	addAccountRoutes(logger, router, accountRepo, transactionRepo)
	addTransactionRoutes(logger, router, accountRepo, transactionRepo)

	// Start business HTTP server
	readTimeout, _ := time.ParseDuration("30s")
	writTimeout, _ := time.ParseDuration("30s")
	idleTimeout, _ := time.ParseDuration("60s")

	serve := &http.Server{
		Addr:    *httpAddr,
		Handler: router,
		TLSConfig: &tls.Config{
			InsecureSkipVerify:       false,
			PreferServerCipherSuites: true,
			MinVersion:               tls.VersionTLS12,
		},
		ReadTimeout:  readTimeout,
		WriteTimeout: writTimeout,
		IdleTimeout:  idleTimeout,
	}
	shutdownServer := func() {
		if err := serve.Shutdown(context.TODO()); err != nil {
			logger.Log("shutdown", err)
		}
	}

	// Start business logic HTTP server
	go func() {
		logger.Log("transport", "HTTP", "addr", *httpAddr)
		errs <- serve.ListenAndServe()
		// TODO(adam): support TLS
		// func (srv *Server) ListenAndServeTLS(certFile, keyFile string) error
	}()

	// Block/Wait for an error
	if err := <-errs; err != nil {
		shutdownServer()
		logger.Log("exit", err)
	}
}

func addPingRoute(logger log.Logger, r *mux.Router) {
	r.Methods("GET").Path("/ping").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if requestId := moovhttp.GetRequestId(r); requestId != "" {
			logger.Log("route", "ping", "requestId", requestId)
		}
		moovhttp.SetAccessControlAllowHeaders(w, r.Header.Get("Origin"))
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("PONG"))
	})
}
