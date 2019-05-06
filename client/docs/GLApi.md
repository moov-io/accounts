# \GLApi

All URIs are relative to *http://localhost:8085*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateAccount**](GLApi.md#CreateAccount) | **Post** /customers/{customer_id}/accounts | Create a new account for a Customer
[**CreateCustomer**](GLApi.md#CreateCustomer) | **Post** /customers | Create a new customer
[**CreateTransaction**](GLApi.md#CreateTransaction) | **Post** /accounts/transactions | Post a transaction against multiple accounts. All transaction lines must sum to zero. No money is created or destroyed in a transaction - only moved from account to account. Accounts can be referred to in a Transaction without creating them first.
[**GetAccountTransactions**](GLApi.md#GetAccountTransactions) | **Get** /accounts/{account_id}/transactions | Get transactions for an account. Ordered descending from their posted date.
[**GetAccountsByCustomerID**](GLApi.md#GetAccountsByCustomerID) | **Get** /customers/{customer_id}/accounts | Retrieves a list of accounts associated with the customer ID.
[**GetGLCustomer**](GLApi.md#GetGLCustomer) | **Get** /customers/{customer_id} | Retrieves a Customer object associated with the customer ID.
[**Ping**](GLApi.md#Ping) | **Get** /ping | Ping the GL service to check if running
[**SearchAccounts**](GLApi.md#SearchAccounts) | **Get** /accounts/search | Search for account which matches all query parameters



## CreateAccount

> Account CreateAccount(ctx, customerId, xUserId, createAccount, optional)
Create a new account for a Customer

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**customerId** | **string**| Customer Id | 
**xUserId** | **string**| Moov User ID header, required in all requests | 
**createAccount** | [**CreateAccount**](CreateAccount.md)|  | 
 **optional** | ***CreateAccountOpts** | optional parameters | nil if no parameters

### Optional Parameters

Optional parameters are passed through a pointer to a CreateAccountOpts struct


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



 **xRequestId** | **optional.String**| Optional Request ID allows application developer to trace requests through the systems logs | 

### Return type

[**Account**](Account.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## CreateCustomer

> CreateCustomer CreateCustomer(ctx, xUserId, optional)
Create a new customer

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**xUserId** | **string**| Moov User ID header, required in all requests | 
 **optional** | ***CreateCustomerOpts** | optional parameters | nil if no parameters

### Optional Parameters

Optional parameters are passed through a pointer to a CreateCustomerOpts struct


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **xRequestId** | **optional.String**| Optional Request ID allows application developer to trace requests through the systems logs | 

### Return type

[**CreateCustomer**](CreateCustomer.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## CreateTransaction

> Transaction CreateTransaction(ctx, xUserId, createTransaction, optional)
Post a transaction against multiple accounts. All transaction lines must sum to zero. No money is created or destroyed in a transaction - only moved from account to account. Accounts can be referred to in a Transaction without creating them first.

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**xUserId** | **string**| Moov User ID header, required in all requests | 
**createTransaction** | [**CreateTransaction**](CreateTransaction.md)|  | 
 **optional** | ***CreateTransactionOpts** | optional parameters | nil if no parameters

### Optional Parameters

Optional parameters are passed through a pointer to a CreateTransactionOpts struct


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **xRequestId** | **optional.String**| Optional Request ID allows application developer to trace requests through the systems logs | 

### Return type

[**Transaction**](Transaction.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetAccountTransactions

> []Transaction GetAccountTransactions(ctx, accountId, xUserId, optional)
Get transactions for an account. Ordered descending from their posted date.

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**accountId** | **string**| Account ID | 
**xUserId** | **string**| Moov User ID header, required in all requests | 
 **optional** | ***GetAccountTransactionsOpts** | optional parameters | nil if no parameters

### Optional Parameters

Optional parameters are passed through a pointer to a GetAccountTransactionsOpts struct


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **limit** | **optional.Float32**| Maximum number of transactions to return | 
 **xRequestId** | **optional.String**| Optional Request ID allows application developer to trace requests through the systems logs | 

### Return type

[**[]Transaction**](Transaction.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetAccountsByCustomerID

> []Account GetAccountsByCustomerID(ctx, customerId, xUserId, optional)
Retrieves a list of accounts associated with the customer ID.

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**customerId** | **string**| Customer Id | 
**xUserId** | **string**| Moov User ID header, required in all requests | 
 **optional** | ***GetAccountsByCustomerIDOpts** | optional parameters | nil if no parameters

### Optional Parameters

Optional parameters are passed through a pointer to a GetAccountsByCustomerIDOpts struct


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **xRequestId** | **optional.String**| Optional Request ID allows application developer to trace requests through the systems logs | 

### Return type

[**[]Account**](Account.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetGLCustomer

> Customer GetGLCustomer(ctx, customerId, xUserId, optional)
Retrieves a Customer object associated with the customer ID.

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**customerId** | **string**| Customer Id | 
**xUserId** | **string**| Moov User ID header, required in all requests | 
 **optional** | ***GetGLCustomerOpts** | optional parameters | nil if no parameters

### Optional Parameters

Optional parameters are passed through a pointer to a GetGLCustomerOpts struct


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **xRequestId** | **optional.String**| Optional Request ID allows application developer to trace requests through the systems logs | 

### Return type

[**Customer**](Customer.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## Ping

> Ping(ctx, )
Ping the GL service to check if running

### Required Parameters

This endpoint does not need any parameter.

### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SearchAccounts

> Account SearchAccounts(ctx, number, routingNumber, type_, xUserId, optional)
Search for account which matches all query parameters

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**number** | **string**| Account number | 
**routingNumber** | **string**| ABA routing number for the Financial Institution | 
**type_** | **string**| Account type | 
**xUserId** | **string**| Moov User ID header, required in all requests | 
 **optional** | ***SearchAccountsOpts** | optional parameters | nil if no parameters

### Optional Parameters

Optional parameters are passed through a pointer to a SearchAccountsOpts struct


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------




 **xRequestId** | **optional.String**| Optional Request ID allows application developer to trace requests through the systems logs | 

### Return type

[**Account**](Account.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

