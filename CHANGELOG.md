## v0.3.0 (Unreleased)

BREAKING CHANGES

- This project is called `accounts` now and was renamed from `gl`.

ADDITIONS

- cmd/server: add initial transaction storage with QLedger
- cmd/server: use our qledgerDeployment dockertest setup for qledgerAccountRepository tests
- api,client: add route to for getting account transactions, generate client
- cmd/server: accounts: setup sqlite persistence for gl.Accounts
- cmd/server: transactions: setup initial sqlite persistence for transactions
- cmd/server: add 'balance' to account creation to track initial balance
- cmd/server: compute overdraft correctly by rejecting accounts in the negative
- all: add customerId as query parameter to account search

BUG FIXES

- api: Fix broken OpenAPI Go client generation
- cmd/server: add missing database/sql Rows.Close()
- cmd/server: return database/sql Rows.Err
- cmd/server: accounts: case-insensitive compare for account type

## v0.2.2 (Released 2019-03-27)

BUG FIXES

- build: Switch to Docker image with CGO (for SQLite)

## v0.2.1 (Released 2019-03-27)

BREAKING CHANGES

- client: rename getCustomer to getGLCustomer (for larger api and go-client)

## v0.2.0 (Released 2019-03-26)

ADDITIONS

- cmd/server: Add customer creation route (`POST /customers`)
- cmd/server: Add sqlite and QLedger persistence for accounts and customers

IMPROVEMENTS

- cmd/server: panic if ABA default routing number is empty
- cmd/server: Log errors when making QLedger calls

## v0.1.0 (Released 2019-03-20)

- Initial release
