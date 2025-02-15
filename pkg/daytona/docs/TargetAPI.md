# \TargetAPI

All URIs are relative to *http://localhost:3986*

Method | HTTP request | Description
------------- | ------------- | -------------
[**ListTargets**](TargetAPI.md#ListTargets) | **Get** /target | List targets
[**RemoveTarget**](TargetAPI.md#RemoveTarget) | **Delete** /target/{target} | Remove a target
[**SetDefaultTarget**](TargetAPI.md#SetDefaultTarget) | **Patch** /target/{target}/set-default | Set target to default
[**SetTarget**](TargetAPI.md#SetTarget) | **Put** /target | Set a target



## ListTargets

> []ProviderTarget ListTargets(ctx).Execute()

List targets



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/hide-org/hide"
)

func main() {

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.TargetAPI.ListTargets(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `TargetAPI.ListTargets``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ListTargets`: []ProviderTarget
	fmt.Fprintf(os.Stdout, "Response from `TargetAPI.ListTargets`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiListTargetsRequest struct via the builder pattern


### Return type

[**[]ProviderTarget**](ProviderTarget.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## RemoveTarget

> RemoveTarget(ctx, target).Execute()

Remove a target



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/hide-org/hide"
)

func main() {
	target := "target_example" // string | Target name

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.TargetAPI.RemoveTarget(context.Background(), target).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `TargetAPI.RemoveTarget``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**target** | **string** | Target name | 

### Other Parameters

Other parameters are passed through a pointer to a apiRemoveTargetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

 (empty response body)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SetDefaultTarget

> SetDefaultTarget(ctx, target).Execute()

Set target to default



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/hide-org/hide"
)

func main() {
	target := "target_example" // string | Target name

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.TargetAPI.SetDefaultTarget(context.Background(), target).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `TargetAPI.SetDefaultTarget``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**target** | **string** | Target name | 

### Other Parameters

Other parameters are passed through a pointer to a apiSetDefaultTargetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

 (empty response body)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SetTarget

> SetTarget(ctx).Target(target).Execute()

Set a target



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/hide-org/hide"
)

func main() {
	target := *openapiclient.NewCreateProviderTargetDTO("Name_example", "Options_example", *openapiclient.NewProviderProviderInfo("Name_example", "Version_example")) // CreateProviderTargetDTO | Target to set

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.TargetAPI.SetTarget(context.Background()).Target(target).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `TargetAPI.SetTarget``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiSetTargetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **target** | [**CreateProviderTargetDTO**](CreateProviderTargetDTO.md) | Target to set | 

### Return type

 (empty response body)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

