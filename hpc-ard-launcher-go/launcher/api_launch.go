/*
Launcher API

The Launcher API is the execution layer for the Capsules framework.  It handles all the details of launching and monitoring runtime environments.

API version: 3.2.9
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package launcher

import (
	"bytes"
	_context "context"
	_ioutil "io/ioutil"
	_nethttp "net/http"
	_neturl "net/url"
	"strings"
	"os"
)

// Linger please
var (
	_ _context.Context
)

// LaunchApiService LaunchApi service
type LaunchApiService service

type ApiAddCredentialRequest struct {
	ctx _context.Context
	ApiService *LaunchApiService
	owner string
	name string
	body **os.File
}

// The credential data to store
func (r ApiAddCredentialRequest) Body(body *os.File) ApiAddCredentialRequest {
	r.body = &body
	return r
}

func (r ApiAddCredentialRequest) Execute() (*_nethttp.Response, error) {
	return r.ApiService.AddCredentialExecute(r)
}

/*
AddCredential Creates/updates a credential that the Dispatch Centre can use to launch environments on behalf of the user

 @param ctx _context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 @param owner The username of the user whose resources that you wish to access
 @param name The name of the resource that you wish to access
 @return ApiAddCredentialRequest
*/
func (a *LaunchApiService) AddCredential(ctx _context.Context, owner string, name string) ApiAddCredentialRequest {
	return ApiAddCredentialRequest{
		ApiService: a,
		ctx: ctx,
		owner: owner,
		name: name,
	}
}

// Execute executes the request
func (a *LaunchApiService) AddCredentialExecute(r ApiAddCredentialRequest) (*_nethttp.Response, error) {
	var (
		localVarHTTPMethod   = _nethttp.MethodPut
		localVarPostBody     interface{}
		formFiles            []formFile
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "LaunchApiService.AddCredential")
	if err != nil {
		return nil, GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/launch/credentials/{owner}/{name}"
	localVarPath = strings.Replace(localVarPath, "{"+"owner"+"}", _neturl.PathEscape(parameterToString(r.owner, "")), -1)
	localVarPath = strings.Replace(localVarPath, "{"+"name"+"}", _neturl.PathEscape(parameterToString(r.name, "")), -1)

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := _neturl.Values{}
	localVarFormParams := _neturl.Values{}
	if r.body == nil {
		return nil, reportError("body is required and must be specified")
	}

	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{"application/octet-stream"}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	// body params
	localVarPostBody = r.body
	req, err := a.client.prepareRequest(r.ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, formFiles)
	if err != nil {
		return nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(req)
	if err != nil || localVarHTTPResponse == nil {
		return localVarHTTPResponse, err
	}

	localVarBody, err := _ioutil.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	localVarHTTPResponse.Body = _ioutil.NopCloser(bytes.NewBuffer(localVarBody))
	if err != nil {
		return localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
		return localVarHTTPResponse, newErr
	}

	return localVarHTTPResponse, nil
}

type ApiHasCredentialRequest struct {
	ctx _context.Context
	ApiService *LaunchApiService
	owner string
	name string
}


func (r ApiHasCredentialRequest) Execute() (*_nethttp.Response, error) {
	return r.ApiService.HasCredentialExecute(r)
}

/*
HasCredential Determines whether a given credential has been provided

 @param ctx _context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 @param owner The username of the user whose resources that you wish to access
 @param name The name of the resource that you wish to access
 @return ApiHasCredentialRequest
*/
func (a *LaunchApiService) HasCredential(ctx _context.Context, owner string, name string) ApiHasCredentialRequest {
	return ApiHasCredentialRequest{
		ApiService: a,
		ctx: ctx,
		owner: owner,
		name: name,
	}
}

// Execute executes the request
func (a *LaunchApiService) HasCredentialExecute(r ApiHasCredentialRequest) (*_nethttp.Response, error) {
	var (
		localVarHTTPMethod   = _nethttp.MethodHead
		localVarPostBody     interface{}
		formFiles            []formFile
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "LaunchApiService.HasCredential")
	if err != nil {
		return nil, GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/launch/credentials/{owner}/{name}"
	localVarPath = strings.Replace(localVarPath, "{"+"owner"+"}", _neturl.PathEscape(parameterToString(r.owner, "")), -1)
	localVarPath = strings.Replace(localVarPath, "{"+"name"+"}", _neturl.PathEscape(parameterToString(r.name, "")), -1)

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := _neturl.Values{}
	localVarFormParams := _neturl.Values{}

	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	req, err := a.client.prepareRequest(r.ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, formFiles)
	if err != nil {
		return nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(req)
	if err != nil || localVarHTTPResponse == nil {
		return localVarHTTPResponse, err
	}

	localVarBody, err := _ioutil.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	localVarHTTPResponse.Body = _ioutil.NopCloser(bytes.NewBuffer(localVarBody))
	if err != nil {
		return localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
		return localVarHTTPResponse, newErr
	}

	return localVarHTTPResponse, nil
}

type ApiLaunchRequest struct {
	ctx _context.Context
	ApiService *LaunchApiService
	manifest *Manifest
	impersonate *string
	dispatchId *string
}

// The manifest to launch
func (r ApiLaunchRequest) Manifest(manifest Manifest) ApiLaunchRequest {
	r.manifest = &manifest
	return r
}
// User to impersonate (user encoded in authorization token must be configured as an administrator)
func (r ApiLaunchRequest) Impersonate(impersonate string) ApiLaunchRequest {
	r.impersonate = &impersonate
	return r
}
// Force the use of a specific DispatchID instead of generation of a new one.
func (r ApiLaunchRequest) DispatchId(dispatchId string) ApiLaunchRequest {
	r.dispatchId = &dispatchId
	return r
}

func (r ApiLaunchRequest) Execute() (DispatchInfo, *_nethttp.Response, error) {
	return r.ApiService.LaunchExecute(r)
}

/*
Launch Launches the runtime environment described by the provided manifest in a synchronous manner

 @param ctx _context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 @return ApiLaunchRequest
*/
func (a *LaunchApiService) Launch(ctx _context.Context) ApiLaunchRequest {
	return ApiLaunchRequest{
		ApiService: a,
		ctx: ctx,
	}
}

// Execute executes the request
//  @return DispatchInfo
func (a *LaunchApiService) LaunchExecute(r ApiLaunchRequest) (DispatchInfo, *_nethttp.Response, error) {
	var (
		localVarHTTPMethod   = _nethttp.MethodPut
		localVarPostBody     interface{}
		formFiles            []formFile
		localVarReturnValue  DispatchInfo
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "LaunchApiService.Launch")
	if err != nil {
		return localVarReturnValue, nil, GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/launch"

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := _neturl.Values{}
	localVarFormParams := _neturl.Values{}
	if r.manifest == nil {
		return localVarReturnValue, nil, reportError("manifest is required and must be specified")
	}

	if r.impersonate != nil {
		localVarQueryParams.Add("impersonate", parameterToString(*r.impersonate, ""))
	}
	if r.dispatchId != nil {
		localVarQueryParams.Add("dispatchId", parameterToString(*r.dispatchId, ""))
	}
	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{"application/json", "application/yaml"}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{"application/json", "application/yaml"}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	// body params
	localVarPostBody = r.manifest
	req, err := a.client.prepareRequest(r.ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, formFiles)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(req)
	if err != nil || localVarHTTPResponse == nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	localVarBody, err := _ioutil.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	localVarHTTPResponse.Body = _ioutil.NopCloser(bytes.NewBuffer(localVarBody))
	if err != nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	err = a.client.decode(&localVarReturnValue, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
	if err != nil {
		newErr := GenericOpenAPIError{
			body:  localVarBody,
			error: err.Error(),
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	return localVarReturnValue, localVarHTTPResponse, nil
}

type ApiLaunchAsyncRequest struct {
	ctx _context.Context
	ApiService *LaunchApiService
	manifest *Manifest
	impersonate *string
	dispatchId *string
}

// The manifest to launch
func (r ApiLaunchAsyncRequest) Manifest(manifest Manifest) ApiLaunchAsyncRequest {
	r.manifest = &manifest
	return r
}
// User to impersonate (user encoded in authorization token must be configured as an administrator)
func (r ApiLaunchAsyncRequest) Impersonate(impersonate string) ApiLaunchAsyncRequest {
	r.impersonate = &impersonate
	return r
}
// Force the use of a specific DispatchID instead of generation of a new one.
func (r ApiLaunchAsyncRequest) DispatchId(dispatchId string) ApiLaunchAsyncRequest {
	r.dispatchId = &dispatchId
	return r
}

func (r ApiLaunchAsyncRequest) Execute() (DispatchInfo, *_nethttp.Response, error) {
	return r.ApiService.LaunchAsyncExecute(r)
}

/*
LaunchAsync Launches the runtime environment described by the provided manifest in an asynchronous manner

 @param ctx _context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 @return ApiLaunchAsyncRequest
*/
func (a *LaunchApiService) LaunchAsync(ctx _context.Context) ApiLaunchAsyncRequest {
	return ApiLaunchAsyncRequest{
		ApiService: a,
		ctx: ctx,
	}
}

// Execute executes the request
//  @return DispatchInfo
func (a *LaunchApiService) LaunchAsyncExecute(r ApiLaunchAsyncRequest) (DispatchInfo, *_nethttp.Response, error) {
	var (
		localVarHTTPMethod   = _nethttp.MethodPut
		localVarPostBody     interface{}
		formFiles            []formFile
		localVarReturnValue  DispatchInfo
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "LaunchApiService.LaunchAsync")
	if err != nil {
		return localVarReturnValue, nil, GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/launch/async"

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := _neturl.Values{}
	localVarFormParams := _neturl.Values{}
	if r.manifest == nil {
		return localVarReturnValue, nil, reportError("manifest is required and must be specified")
	}

	if r.impersonate != nil {
		localVarQueryParams.Add("impersonate", parameterToString(*r.impersonate, ""))
	}
	if r.dispatchId != nil {
		localVarQueryParams.Add("dispatchId", parameterToString(*r.dispatchId, ""))
	}
	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{"application/json", "application/yaml"}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{"application/json", "application/yaml"}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	// body params
	localVarPostBody = r.manifest
	req, err := a.client.prepareRequest(r.ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, formFiles)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(req)
	if err != nil || localVarHTTPResponse == nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	localVarBody, err := _ioutil.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	localVarHTTPResponse.Body = _ioutil.NopCloser(bytes.NewBuffer(localVarBody))
	if err != nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	err = a.client.decode(&localVarReturnValue, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
	if err != nil {
		newErr := GenericOpenAPIError{
			body:  localVarBody,
			error: err.Error(),
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	return localVarReturnValue, localVarHTTPResponse, nil
}

type ApiRemoveCredentialRequest struct {
	ctx _context.Context
	ApiService *LaunchApiService
	owner string
	name string
}


func (r ApiRemoveCredentialRequest) Execute() (*_nethttp.Response, error) {
	return r.ApiService.RemoveCredentialExecute(r)
}

/*
RemoveCredential Removes a credential

 @param ctx _context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 @param owner The username of the user whose resources that you wish to access
 @param name The name of the resource that you wish to access
 @return ApiRemoveCredentialRequest
*/
func (a *LaunchApiService) RemoveCredential(ctx _context.Context, owner string, name string) ApiRemoveCredentialRequest {
	return ApiRemoveCredentialRequest{
		ApiService: a,
		ctx: ctx,
		owner: owner,
		name: name,
	}
}

// Execute executes the request
func (a *LaunchApiService) RemoveCredentialExecute(r ApiRemoveCredentialRequest) (*_nethttp.Response, error) {
	var (
		localVarHTTPMethod   = _nethttp.MethodDelete
		localVarPostBody     interface{}
		formFiles            []formFile
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "LaunchApiService.RemoveCredential")
	if err != nil {
		return nil, GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/launch/credentials/{owner}/{name}"
	localVarPath = strings.Replace(localVarPath, "{"+"owner"+"}", _neturl.PathEscape(parameterToString(r.owner, "")), -1)
	localVarPath = strings.Replace(localVarPath, "{"+"name"+"}", _neturl.PathEscape(parameterToString(r.name, "")), -1)

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := _neturl.Values{}
	localVarFormParams := _neturl.Values{}

	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	req, err := a.client.prepareRequest(r.ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, formFiles)
	if err != nil {
		return nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(req)
	if err != nil || localVarHTTPResponse == nil {
		return localVarHTTPResponse, err
	}

	localVarBody, err := _ioutil.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	localVarHTTPResponse.Body = _ioutil.NopCloser(bytes.NewBuffer(localVarBody))
	if err != nil {
		return localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
		return localVarHTTPResponse, newErr
	}

	return localVarHTTPResponse, nil
}
