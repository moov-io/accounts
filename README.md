moov-io/accounts
===

[![GoDoc](https://godoc.org/github.com/moov-io/accounts?status.svg)](https://godoc.org/github.com/moov-io/accounts)
[![Build Status](https://travis-ci.com/moov-io/accounts.svg?branch=master)](https://travis-ci.com/moov-io/accounts)
[![Coverage Status](https://codecov.io/gh/moov-io/accounts/branch/master/graph/badge.svg)](https://codecov.io/gh/moov-io/accounts)
[![Go Report Card](https://goreportcard.com/badge/github.com/moov-io/accounts)](https://goreportcard.com/report/github.com/moov-io/accounts)
[![Apache 2 licensed](https://img.shields.io/badge/license-Apache2-blue.svg)](https://raw.githubusercontent.com/moov-io/accounts/master/LICENSE)

Moov Accounts is a [general ledger](https://en.wikipedia.org/wiki/General_ledger) accounting system designed to support the handling of Customer funds deposited at a bank or credit union. Implemented as an RESTful API and Moov Accounts can be leveraged by a financial institution to provide modern banking services to its customers. Moov Accounts can be utilized by a technology company to manage Customer assets that are in a single For Benefit of Account at a financial institution. Moov's primary use is with [paygate](https://github.com/moov-io/paygate). (A full implementation of ACH origination)

*This implementation is currently not complete for use in production, but any advice or feedback would be greatly appreciated!*

Docs: [docs.moov.io](https://docs.moov.io/) | [api docs](https://api.moov.io/apps/accounts/)

### Reading

Accounting for Developers [Part 1](https://docs.google.com/document/d/1HDLRa6vKpclO1JtxbGB5NeAYWf8cf1UMGy22o8OZZq4/edit#heading=h.jo5avukxj1q), [Part 2](https://docs.google.com/document/d/1qhtirHUzPu7Od7yX3A4kA424tjFCv5Kbi42xj49tKlw/edit), [Part 3](https://docs.google.com/document/d/1kIwonczHvJLgzcijLtljHc5fccQ6fKI6TodhnGYHCEA/edit).

### Deployment

You can download [our docker image `moov/accounts`](https://hub.docker.com/r/moov/accounts/) from Docker Hub or use this repository. No configuration is required to serve on `:8085` and metrics at `:9095/metrics` in Prometheus format.

### Configuration

The following environmental variables can be set to configure behavior in Accounts.

| Environmental Variable | Description | Default |
|-----|-----|-----|
| `DEFAULT_ROUTING_NUMBER` | ABA routing number used when accounts are created. | Required |
| `SQLITE_DB_PATH`| Local filepath location for the paygate SQLite database. | `accounts.db` |
| `ACCOUNT_STORAGE_TYPE` | Storage engine for account data. | Default: `qledger` |
| `TRANSACTION_STORAGE_TYPE` | Storage engine for transaction data. | Default: `qledger` |
| `QLEDGER_ENDPOINT` | HTTP endpoint to access QLedger (if storage type is `qledger`) | Required |
| `QLEDGER_AUTH_TOKEN` | Auth token to access QLedger (if storage type is `qledger`) | Required |
| `LOG_FORMAT` | Format for logging lines to be written as. | Options: `json`, `plain` - Default: `plain` |
| `HTTP_BIND_ADDRESS` | Address for paygate to bind its HTTP server on. This overrides the command-line flag `-http.addr`. | Default: `:8085` |
| `HTTP_ADMIN_BIND_ADDRESS` | Address for paygate to bind its admin HTTP server on. This overrides the command-line flag `-admin.addr`. | Default: `:9095` |


## Getting Help

 channel | info
 ------- | -------
 [Project Documentation](https://docs.moov.io/) | Our project documentation available online.
 Google Group [moov-users](https://groups.google.com/forum/#!forum/moov-users)| The Moov users Google group is for contributors other people contributing to the Moov project. You can join them without a google account by sending an email to [moov-users+subscribe@googlegroups.com](mailto:moov-users+subscribe@googlegroups.com). After receiving the join-request message, you can simply reply to that to confirm the subscription.
Twitter [@moov_io](https://twitter.com/moov_io)	| You can follow Moov.IO's Twitter feed to get updates on our project(s). You can also tweet us questions or just share blogs or stories.
[GitHub Issue](https://github.com/moov-io) | If you are able to reproduce an problem please open a GitHub Issue under the specific project that caused the error.
[moov-io slack](http://moov-io.slack.com/) | Join our slack channel to have an interactive discussion about the development of the project. [Request an invite to the slack channel](https://join.slack.com/t/moov-io/shared_invite/enQtNDE5NzIwNTYxODEwLTRkYTcyZDI5ZTlkZWRjMzlhMWVhMGZlOTZiOTk4MmM3MmRhZDY4OTJiMDVjOTE2MGEyNWYzYzY1MGMyMThiZjg)

## Contributing

Yes please! Please review our [Contributing guide](CONTRIBUTING.md) and [Code of Conduct](https://github.com/moov-io/ach/blob/master/CODE_OF_CONDUCT.md) to get started!

Note: This project uses Go Modules, which requires Go 1.11 or higher, but we ship the vendor directory in our repository.

## License

Apache License 2.0 See [LICENSE](LICENSE) for details.
