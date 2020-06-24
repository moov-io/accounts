This guide is a walkthrough for using Moov Accounts.

Refer to the [Accounts API documentation](https://moov-io.github.io/accounts/) for additional details and an OpenAPI specification of these endpoints.

### Create an Account

```
$ curl -XPOST -H "x-user-id: 8ebaf4dc" --data '{
  "customerID": "0ecf8f9a",
  "balance": 100,
  "name": "Basic Checking",
  "type": "Checking"
}' http://localhost:8085/accounts | jq .

```

Returns

```
{
  "ID": "ed4c07e1255e84bddf473e4ad4082f763e02e110",
  "customerID": "foo",
  "name": "Basic Checking",
  "accountNumber": "909481572",
  "routingNumber": "987654320",
  "status": "open",
  "type": "Checking",
  "createdAt": "2020-03-12T20:04:37.282515-04:00",
  "closedAt": "0001-01-01T00:00:00Z",
  "lastModified": "2020-03-12T20:04:37.282515-04:00"
}
```

### Search for an account

```
$ curl -H "x-user-id: 9b80698a" "http://localhost:8085/accounts/search"
```

Query Parameters:

- `number`: Full account number
- `routingNumber`: Valid ABA routing number
- `type`: `Checking` or `Savings`
- `customerID`: Random UUID used to assign the account with a Customer

### Post transactions against accounts

```
$ curl -XPOST -H "x-user-id: 5134a914" --data '{
    "lines": [
        {
            "id":  "",
            "timestamp": "2020-01-02T15:04:05Z07:00",
            "lines": [
                {
                    "accountID": "d487bca5",
                    "purpose": "ACHDebit",
                    "amount": 1277
                },
                {
                    "accountID": "70b3dde7",
                    "purpose": "ACHCredit",
                    "amount": 1277
                },
            ]
        }
    ]
}' http://localhost:8085/accounts/transactions | jq .
```

### Retrieve an account transaction

```
$ curl -H "x-user-id: 8f0eafba" http://localhost:8085/accounts/{accountID}/transactions
```
