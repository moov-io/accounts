# \AccountsApi

All URIs are relative to *http://localhost:8085*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateAccount**](AccountsApi.md#CreateAccount) | **Post** /accounts | Create a new account for a Customer
[**CreateTransaction**](AccountsApi.md#CreateTransaction) | **Post** /accounts/transactions | Post a transaction against multiple accounts. All transaction lines must sum to zero. No money is created or destroyed in a transaction - only moved from account to account. Accounts can be referred to in a Transaction without creating them first.
[**GetAccountTransactions**](AccountsApi.md#GetAccountTransactions) | **Get** /accounts/{accountID}/transactions | Get transactions for an account. Ordered descending from their posted date.
[**Ping**](AccountsApi.md#Ping) | **Get** /ping | Ping the Accounts service to check if running
[**ReverseTransaction**](AccountsApi.md#ReverseTransaction) | **Post** /accounts/transactions/{transaction_id}/reversal | Reverse a transaction by debiting the credited and crediting the debited amounts among all accounts involved.
[**SearchAccounts**](AccountsApi.md#SearchAccounts) | **Get** /accounts/search | Search for account which matches all query parameters



## CreateAccount

> Account CreateAccount(ctx, xUserID, createAccount, optional)
Create a new account for a Customer

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**xUserID** | **string**| Moov User ID header, required in all requests | 
**createAccount** | [**CreateAccount**](CreateAccount.md)|  | 
 **optional** | ***CreateAccountOpts** | optional parameters | nil if no parameters

### Optional Parameters

Optional parameters are passed through a pointer to a CreateAccountOpts struct


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **xRequestID** | **optional.String**| Optional Request ID allows application developer to trace requests through the systems logs | 

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


## CreateTransaction

> Transaction CreateTransaction(ctx, xUserID, createTransaction, optional)
Post a transaction against multiple accounts. All transaction lines must sum to zero. No money is created or destroyed in a transaction - only moved from account to account. Accounts can be referred to in a Transaction without creating them first.

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**xUserID** | **string**| Moov User ID header, required in all requests | 
**createTransaction** | [**CreateTransaction**](CreateTransaction.md)|  | 
 **optional** | ***CreateTransactionOpts** | optional parameters | nil if no parameters

### Optional Parameters

Optional parameters are passed through a pointer to a CreateTransactionOpts struct


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **xRequestID** | **optional.String**| Optional Request ID allows application developer to trace requests through the systems logs | 

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

> []Transaction GetAccountTransactions(ctx, accountID, xUserID, optional)
Get transactions for an account. Ordered descending from their posted date.

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**accountID** | **string**| Account ID | 
**xUserID** | **string**| Moov User ID header, required in all requests | 
 **optional** | ***GetAccountTransactionsOpts** | optional parameters | nil if no parameters

### Optional Parameters

Optional parameters are passed through a pointer to a GetAccountTransactionsOpts struct


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **limit** | **optional.Float32**| Maximum number of transactions to return | 
 **xRequestID** | **optional.String**| Optional Request ID allows application developer to trace requests through the systems logs | 

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


## Ping

> Ping(ctx, )
Ping the Accounts service to check if running

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


## ReverseTransaction

> Transaction ReverseTransaction(ctx, transactionId, xUserID, optional)
Reverse a transaction by debiting the credited and crediting the debited amounts among all accounts involved.

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**transactionId** | **string**| Transaction ID | 
**xUserID** | **string**| Moov User ID header, required in all requests | 
 **optional** | ***ReverseTransactionOpts** | optional parameters | nil if no parameters

### Optional Parameters

Optional parameters are passed through a pointer to a ReverseTransactionOpts struct


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **xRequestID** | **optional.String**| Optional Request ID allows application developer to trace requests through the systems logs | 

### Return type

[**Transaction**](Transaction.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SearchAccounts

> []Account SearchAccounts(ctx, xUserID, optional)
Search for account which matches all query parameters

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**xUserID** | **string**| Moov User ID header, required in all requests | 
 **optional** | ***SearchAccountsOpts** | optional parameters | nil if no parameters

### Optional Parameters

Optional parameters are passed through a pointer to a SearchAccountsOpts struct


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **number** | **optional.String**| Account number | 
 **routingNumber** | **optional.String**| ABA routing number for the Financial Institution | 
 **type_** | **optional.String**| Account type | 
 **customerID** | **optional.String**| Customer ID associated to accounts | 
 **xRequestID** | **optional.String**| Optional Request ID allows application developer to trace requests through the systems logs | 

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

