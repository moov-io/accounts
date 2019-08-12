# Account

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **string** | The unique identifier for an account | [optional] 
**CustomerID** | **string** | The unique identifier for the customer who owns the account | [optional] 
**Name** | **string** | Caller defined label for this account. | [optional] 
**AccountNumber** | **string** | A unique Account number at the bank. | [optional] 
**AccountNumberMasked** | **string** | Last four digits of an account number | [optional] 
**RoutingNumber** | **string** | Routing Transit Number is a nine-digit number assigned by the ABA | [optional] 
**Status** | **string** | Status of the account being created. | [optional] 
**Type** | **string** | Product type of the account | [optional] 
**CreatedAt** | [**time.Time**](time.Time.md) |  | [optional] 
**ClosedAt** | [**time.Time**](time.Time.md) |  | [optional] 
**LastModified** | [**time.Time**](time.Time.md) | Last time the object was modified except balances | [optional] 
**Balance** | **int32** | Total balance of account in USD cents. | [optional] 
**BalanceAvailable** | **int32** | Balance available in USD cents to be drawn | [optional] 
**BalancePending** | **int32** | Balance of pending transactions in USD cents | [optional] 

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


