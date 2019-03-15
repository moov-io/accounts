# \GLApi

All URIs are relative to *http://localhost:8085*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateAccount**](GLApi.md#CreateAccount) | **Post** /customers/{customer_id}/accounts | Create a new account for a Customer
[**GetAccountsByCustomerID**](GLApi.md#GetAccountsByCustomerID) | **Get** /customers/{customer_id}/accounts | Retrieves a list of accounts associated with the customer ID.
[**GetCustomer**](GLApi.md#GetCustomer) | **Get** /customers/{customer_id} | Retrieves a Customer object associated with the customer ID.
[**Ping**](GLApi.md#Ping) | **Get** /ping | Ping the GL service to check if running


# **CreateAccount**
> Account CreateAccount(ctx, customerId, createAccount)
Create a new account for a Customer

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **customerId** | **string**| Customer Id | 
  **createAccount** | [**CreateAccount**](CreateAccount.md)|  | 

### Return type

[**Account**](Account.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **GetAccountsByCustomerID**
> []Account GetAccountsByCustomerID(ctx, customerId)
Retrieves a list of accounts associated with the customer ID.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **customerId** | **string**| Customer Id | 

### Return type

[**[]Account**](Account.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **GetCustomer**
> Customer GetCustomer(ctx, customerId)
Retrieves a Customer object associated with the customer ID.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **customerId** | **string**| Customer Id | 

### Return type

[**Customer**](Customer.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **Ping**
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

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

