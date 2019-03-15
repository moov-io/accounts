# \AccountApi

All URIs are relative to *http://localhost:8085*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateAccount**](AccountApi.md#CreateAccount) | **Post** /customers/{customer_id}/accounts | Create a new account for a Customer
[**GetAccountsByCustomerID**](AccountApi.md#GetAccountsByCustomerID) | **Get** /customers/{customer_id}/accounts | Retrieves a list of accounts associated with the customer ID.


# **CreateAccount**
> InlineResponse200 CreateAccount(ctx, customerId, createAccount)
Create a new account for a Customer

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **customerId** | **string**| Customer Id | 
  **createAccount** | [**CreateAccount**](CreateAccount.md)|  | 

### Return type

[**InlineResponse200**](inline_response_200.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: text/plain, application/json

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

