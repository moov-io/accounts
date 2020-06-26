# Overview

<a class="github-button" href="https://github.com/moov-io/accounts" data-size="large" data-show-count="true" aria-label="Star moov-io/accounts on GitHub">moov-io/accounts</a>
<a href="https://godoc.org/github.com/moov-io/accounts/client"><img src="https://godoc.org/github.com/moov-io/accounts/client?status.svg" /></a>
<a href="https://raw.githubusercontent.com/moov-io/accounts/master/LICENSE"><img src="https://img.shields.io/badge/license-Apache2-blue.svg" /></a>

Moov Accounts is a [general ledger](https://en.wikipedia.org/wiki/General_ledger) accounting system designed to support the handling of Customer funds deposited at a bank or credit union. Implemented as an RESTful API and Moov Accounts can be leveraged by a financial institution to provide modern banking services to its customers. Moov Accounts can be utilized by a technology company to manage Customer assets that are in a single For Benefit of Account at a financial institution. Moov's primary use is with [PayGate](https://github.com/moov-io/paygate). (A full implementation of ACH origination)

Docs: [Project](https://moov-io.github.io/accounts/) | [API Endpoints](https://moov-io.github.io/accounts/api/)

## Running Moov Accounts Server

- <a href="#binary-distribution">Binary Distributions</a> are released with every versioned release. Frequently added to the VM/AMI build script for the application needing Moov Accounts.
- A <a href="#docker-container">Docker container</a> is built and added to Docker Hub with every versioned released.
- Our hosted [api.moov.io](https://api.moov.io) is updated with every versioned release. Our Kubernetes example is what Moov utilizes in our production environment.

### Binary Distribution

Download the [latest Moov Accounts server release](https://github.com/moov-io/accounts/releases) for your operating system and run it from a terminal.

```sh
$ DEFAULT_ROUTING_NUMBER=987654320 ./accounts-darwin-amd64
ts=2020-03-12T23:58:00.457866Z caller=main.go:51 main="Starting moov/accounts server version v0.4.0"
ts=2020-03-12T23:58:00.458897Z caller=database.go:20 database="looking for sqlite database provider"
ts=2020-03-12T23:58:00.458967Z caller=sqlite.go:69 main="sqlite version 3.30.1"
ts=2020-03-12T23:58:00.458926Z caller=main.go:80 admin="listening on [::]:9095"
ts=2020-03-12T23:58:00.460028Z caller=main.go:99 main="using *main.sqlAccountRepository for account storage"
ts=2020-03-12T23:58:00.460059Z caller=database.go:20 database="looking for sqlite database provider"
ts=2020-03-12T23:58:00.460536Z caller=main.go:112 main="using *main.sqlTransactionRepository for transaction storage"
ts=2020-03-12T23:58:00.460723Z caller=main.go:158 main="binding to :8085 for HTTP server"
```

### Docker Container

Moov Accounts is dependent on Docker being properly installed and running on your machine. Ensure that Docker is running. If your Docker client has issues connecting to the service review the [Docker getting started guide](https://docs.docker.com/get-started/) if you have any issues.

Execute the Docker run command

```sh
$ docker run -p 8085:8085 -p 9095:9095 moov/accounts:latest
ts=2020-03-12T23:58:00.457866Z caller=main.go:51 main="Starting moov/accounts server version v0.4.0"
ts=2020-03-12T23:58:00.458897Z caller=database.go:20 database="looking for sqlite database provider"
ts=2020-03-12T23:58:00.458967Z caller=sqlite.go:69 main="sqlite version 3.30.1"
ts=2020-03-12T23:58:00.458926Z caller=main.go:80 admin="listening on [::]:9095"
ts=2020-03-12T23:58:00.460028Z caller=main.go:99 main="using *main.sqlAccountRepository for account storage"
ts=2020-03-12T23:58:00.460059Z caller=database.go:20 database="looking for sqlite database provider"
ts=2020-03-12T23:58:00.460536Z caller=main.go:112 main="using *main.sqlTransactionRepository for transaction storage"
ts=2020-03-12T23:58:00.460723Z caller=main.go:158 main="binding to :8085 for HTTP server"
```

### Kubernetes

Moov deploys Accounts from [this manifest template](https://github.com/moov-io/infra/blob/master/lib/apps/14-accounts.yml) on [Kubernetes](https://kubernetes.io/docs/tutorials/kubernetes-basics/) in the `apps` namespace. You could reach the Accounts instance using `http://accounts.apps.svc.cluster.local:8080` inside the cluster.

We also offer a [Helm Chart](https://github.com/moov-io/charts/tree/master/charts) for deployment.

## Configuration

View our [section on environmental variables](https://github.com/moov-io/accounts#configuration) for options that Accounts accepts.

For database storage we offer [SQLite](https://github.com/moov-io/accounts#sqlite) (default) and [MySQL](https://github.com/moov-io/accounts#mysql) (in v0.5.0-dev) with various configuration options.

## Connecting to Moov Accounts

The Moov Accounts service will be running on port `8085` (with an admin port on `9095`).

Confirm that the service is running by issuing the following command or simply visiting the url in your browser [localhost:8085/ping](http://localhost:8085/ping)

```bash
$ curl http://localhost:8085/ping
PONG
```

### Accounts Admin Port

The port `:9095` is bound by Accounts for our admin service. This HTTP server has endpoints for Prometheus metrics (`GET /metrics`), readiness (`GET /ready`) and liveness checks (`GET /live`).

### API documentation

See our [API documentation](https://moov-io.github.io/accounts/api/) for Moov Accounts endpoints.

## Getting Help

 channel | info
 ------- | -------
 [Project Documentation](https://moov-io.github.io/accounts/) | Our project documentation available online.
 Google Group [moov-users](https://groups.google.com/forum/#!forum/moov-users)| The Moov users Google group is for contributors other people contributing to the Moov project. You can join them without a google account by sending an email to [moov-users+subscribe@googlegroups.com](mailto:moov-users+subscribe@googlegroups.com). After receiving the join-request message, you can simply reply to that to confirm the subscription.
Twitter [@moov_io](https://twitter.com/moov_io)	| You can follow Moov.IO's Twitter feed to get updates on our project(s). You can also tweet us questions or just share blogs or stories.
[GitHub Issue](https://github.com/moov-io) | If you are able to reproduce a problem please open a GitHub Issue under the specific project that caused the error.
[moov-io slack](https://slack.moov.io/) | Join our slack channel (`#accounts`) to have an interactive discussion about the development of the project.
