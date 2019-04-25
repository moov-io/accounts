moov-io/gl
===

[![GoDoc](https://godoc.org/github.com/moov-io/gl?status.svg)](https://godoc.org/github.com/moov-io/gl)
[![Build Status](https://travis-ci.com/moov-io/gl.svg?branch=master)](https://travis-ci.com/moov-io/gl)
[![Coverage Status](https://codecov.io/gh/moov-io/gl/branch/master/graph/badge.svg)](https://codecov.io/gh/moov-io/gl)
[![Go Report Card](https://goreportcard.com/badge/github.com/moov-io/gl)](https://goreportcard.com/report/github.com/moov-io/gl)
[![Apache 2 licensed](https://img.shields.io/badge/license-Apache2-blue.svg)](https://raw.githubusercontent.com/moov-io/gl/master/LICENSE)

*project is under active development and is not production ready*

### Reading

Accounting for Developers [Part 1](https://docs.google.com/document/d/1HDLRa6vKpclO1JtxbGB5NeAYWf8cf1UMGy22o8OZZq4/edit#heading=h.jo5avukxj1q), [Part 2](https://docs.google.com/document/d/1qhtirHUzPu7Od7yX3A4kA424tjFCv5Kbi42xj49tKlw/edit), [Part 3](https://docs.google.com/document/d/1kIwonczHvJLgzcijLtljHc5fccQ6fKI6TodhnGYHCEA/edit).

### Configuration

TODO

| Environmental Variable | Description | Default |
|-----|-----|-----|
| `DEFAULT_ROUTING_NUMBER` | ABA routing number used when accounts are created. | Required |
| `SQLITE_DB_PATH`| Local filepath location for the paygate SQLite database. | `ofac.db` |
| `ACCOUNT_STORAGE_TYPE` | Storage engine for account data. | Default: `qledger` |
| `TRANSACTION_STORAGE_TYPE` | Storage engine for transaction data. | Default: `qledger` |
| `QLEDGER_ENDPOINT` | HTTP endpoint to access QLedger (if storage type is `qledger`) | Required |
| `QLEDGER_AUTH_TOKEN` | Auth token to access QLedger (if storage type is `qledger`) | Required |
