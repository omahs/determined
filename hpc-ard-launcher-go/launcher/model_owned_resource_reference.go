/*
Dispatch Centre API

The Dispatch Centre API is the execution layer for the Capsules framework.  It handles all the details of launching and monitoring runtime environments.

API version: 2.7.12
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package launcher

import (
	"encoding/json"
)

// OwnedResourceReference struct for OwnedResourceReference
type OwnedResourceReference struct {
	Owner *string `json:"owner,omitempty"`
	Name *string `json:"name,omitempty"`
	Version *string `json:"version,omitempty"`
}

// NewOwnedResourceReference instantiates a new OwnedResourceReference object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewOwnedResourceReference() *OwnedResourceReference {
	this := OwnedResourceReference{}
	return &this
}

// NewOwnedResourceReferenceWithDefaults instantiates a new OwnedResourceReference object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewOwnedResourceReferenceWithDefaults() *OwnedResourceReference {
	this := OwnedResourceReference{}
	return &this
}

// GetOwner returns the Owner field value if set, zero value otherwise.
func (o *OwnedResourceReference) GetOwner() string {
	if o == nil || o.Owner == nil {
		var ret string
		return ret
	}
	return *o.Owner
}

// GetOwnerOk returns a tuple with the Owner field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *OwnedResourceReference) GetOwnerOk() (*string, bool) {
	if o == nil || o.Owner == nil {
		return nil, false
	}
	return o.Owner, true
}

// HasOwner returns a boolean if a field has been set.
func (o *OwnedResourceReference) HasOwner() bool {
	if o != nil && o.Owner != nil {
		return true
	}

	return false
}

// SetOwner gets a reference to the given string and assigns it to the Owner field.
func (o *OwnedResourceReference) SetOwner(v string) {
	o.Owner = &v
}

// GetName returns the Name field value if set, zero value otherwise.
func (o *OwnedResourceReference) GetName() string {
	if o == nil || o.Name == nil {
		var ret string
		return ret
	}
	return *o.Name
}

// GetNameOk returns a tuple with the Name field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *OwnedResourceReference) GetNameOk() (*string, bool) {
	if o == nil || o.Name == nil {
		return nil, false
	}
	return o.Name, true
}

// HasName returns a boolean if a field has been set.
func (o *OwnedResourceReference) HasName() bool {
	if o != nil && o.Name != nil {
		return true
	}

	return false
}

// SetName gets a reference to the given string and assigns it to the Name field.
func (o *OwnedResourceReference) SetName(v string) {
	o.Name = &v
}

// GetVersion returns the Version field value if set, zero value otherwise.
func (o *OwnedResourceReference) GetVersion() string {
	if o == nil || o.Version == nil {
		var ret string
		return ret
	}
	return *o.Version
}

// GetVersionOk returns a tuple with the Version field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *OwnedResourceReference) GetVersionOk() (*string, bool) {
	if o == nil || o.Version == nil {
		return nil, false
	}
	return o.Version, true
}

// HasVersion returns a boolean if a field has been set.
func (o *OwnedResourceReference) HasVersion() bool {
	if o != nil && o.Version != nil {
		return true
	}

	return false
}

// SetVersion gets a reference to the given string and assigns it to the Version field.
func (o *OwnedResourceReference) SetVersion(v string) {
	o.Version = &v
}

func (o OwnedResourceReference) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if o.Owner != nil {
		toSerialize["owner"] = o.Owner
	}
	if o.Name != nil {
		toSerialize["name"] = o.Name
	}
	if o.Version != nil {
		toSerialize["version"] = o.Version
	}
	return json.Marshal(toSerialize)
}

type NullableOwnedResourceReference struct {
	value *OwnedResourceReference
	isSet bool
}

func (v NullableOwnedResourceReference) Get() *OwnedResourceReference {
	return v.value
}

func (v *NullableOwnedResourceReference) Set(val *OwnedResourceReference) {
	v.value = val
	v.isSet = true
}

func (v NullableOwnedResourceReference) IsSet() bool {
	return v.isSet
}

func (v *NullableOwnedResourceReference) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableOwnedResourceReference(val *OwnedResourceReference) *NullableOwnedResourceReference {
	return &NullableOwnedResourceReference{value: val, isSet: true}
}

func (v NullableOwnedResourceReference) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableOwnedResourceReference) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


