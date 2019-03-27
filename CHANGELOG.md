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
