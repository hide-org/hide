/*
Daytona Server API

Daytona Server API

API version: v0.0.0-dev
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package daytona

import (
	"encoding/json"
	"bytes"
	"fmt"
)

// checks if the CreatePrebuildDTO type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &CreatePrebuildDTO{}

// CreatePrebuildDTO struct for CreatePrebuildDTO
type CreatePrebuildDTO struct {
	Branch *string `json:"branch,omitempty"`
	CommitInterval *int32 `json:"commitInterval,omitempty"`
	Id *string `json:"id,omitempty"`
	Retention int32 `json:"retention"`
	TriggerFiles []string `json:"triggerFiles,omitempty"`
}

type _CreatePrebuildDTO CreatePrebuildDTO

// NewCreatePrebuildDTO instantiates a new CreatePrebuildDTO object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewCreatePrebuildDTO(retention int32) *CreatePrebuildDTO {
	this := CreatePrebuildDTO{}
	this.Retention = retention
	return &this
}

// NewCreatePrebuildDTOWithDefaults instantiates a new CreatePrebuildDTO object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewCreatePrebuildDTOWithDefaults() *CreatePrebuildDTO {
	this := CreatePrebuildDTO{}
	return &this
}

// GetBranch returns the Branch field value if set, zero value otherwise.
func (o *CreatePrebuildDTO) GetBranch() string {
	if o == nil || IsNil(o.Branch) {
		var ret string
		return ret
	}
	return *o.Branch
}

// GetBranchOk returns a tuple with the Branch field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *CreatePrebuildDTO) GetBranchOk() (*string, bool) {
	if o == nil || IsNil(o.Branch) {
		return nil, false
	}
	return o.Branch, true
}

// HasBranch returns a boolean if a field has been set.
func (o *CreatePrebuildDTO) HasBranch() bool {
	if o != nil && !IsNil(o.Branch) {
		return true
	}

	return false
}

// SetBranch gets a reference to the given string and assigns it to the Branch field.
func (o *CreatePrebuildDTO) SetBranch(v string) {
	o.Branch = &v
}

// GetCommitInterval returns the CommitInterval field value if set, zero value otherwise.
func (o *CreatePrebuildDTO) GetCommitInterval() int32 {
	if o == nil || IsNil(o.CommitInterval) {
		var ret int32
		return ret
	}
	return *o.CommitInterval
}

// GetCommitIntervalOk returns a tuple with the CommitInterval field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *CreatePrebuildDTO) GetCommitIntervalOk() (*int32, bool) {
	if o == nil || IsNil(o.CommitInterval) {
		return nil, false
	}
	return o.CommitInterval, true
}

// HasCommitInterval returns a boolean if a field has been set.
func (o *CreatePrebuildDTO) HasCommitInterval() bool {
	if o != nil && !IsNil(o.CommitInterval) {
		return true
	}

	return false
}

// SetCommitInterval gets a reference to the given int32 and assigns it to the CommitInterval field.
func (o *CreatePrebuildDTO) SetCommitInterval(v int32) {
	o.CommitInterval = &v
}

// GetId returns the Id field value if set, zero value otherwise.
func (o *CreatePrebuildDTO) GetId() string {
	if o == nil || IsNil(o.Id) {
		var ret string
		return ret
	}
	return *o.Id
}

// GetIdOk returns a tuple with the Id field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *CreatePrebuildDTO) GetIdOk() (*string, bool) {
	if o == nil || IsNil(o.Id) {
		return nil, false
	}
	return o.Id, true
}

// HasId returns a boolean if a field has been set.
func (o *CreatePrebuildDTO) HasId() bool {
	if o != nil && !IsNil(o.Id) {
		return true
	}

	return false
}

// SetId gets a reference to the given string and assigns it to the Id field.
func (o *CreatePrebuildDTO) SetId(v string) {
	o.Id = &v
}

// GetRetention returns the Retention field value
func (o *CreatePrebuildDTO) GetRetention() int32 {
	if o == nil {
		var ret int32
		return ret
	}

	return o.Retention
}

// GetRetentionOk returns a tuple with the Retention field value
// and a boolean to check if the value has been set.
func (o *CreatePrebuildDTO) GetRetentionOk() (*int32, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Retention, true
}

// SetRetention sets field value
func (o *CreatePrebuildDTO) SetRetention(v int32) {
	o.Retention = v
}

// GetTriggerFiles returns the TriggerFiles field value if set, zero value otherwise.
func (o *CreatePrebuildDTO) GetTriggerFiles() []string {
	if o == nil || IsNil(o.TriggerFiles) {
		var ret []string
		return ret
	}
	return o.TriggerFiles
}

// GetTriggerFilesOk returns a tuple with the TriggerFiles field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *CreatePrebuildDTO) GetTriggerFilesOk() ([]string, bool) {
	if o == nil || IsNil(o.TriggerFiles) {
		return nil, false
	}
	return o.TriggerFiles, true
}

// HasTriggerFiles returns a boolean if a field has been set.
func (o *CreatePrebuildDTO) HasTriggerFiles() bool {
	if o != nil && !IsNil(o.TriggerFiles) {
		return true
	}

	return false
}

// SetTriggerFiles gets a reference to the given []string and assigns it to the TriggerFiles field.
func (o *CreatePrebuildDTO) SetTriggerFiles(v []string) {
	o.TriggerFiles = v
}

func (o CreatePrebuildDTO) MarshalJSON() ([]byte, error) {
	toSerialize,err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o CreatePrebuildDTO) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	if !IsNil(o.Branch) {
		toSerialize["branch"] = o.Branch
	}
	if !IsNil(o.CommitInterval) {
		toSerialize["commitInterval"] = o.CommitInterval
	}
	if !IsNil(o.Id) {
		toSerialize["id"] = o.Id
	}
	toSerialize["retention"] = o.Retention
	if !IsNil(o.TriggerFiles) {
		toSerialize["triggerFiles"] = o.TriggerFiles
	}
	return toSerialize, nil
}

func (o *CreatePrebuildDTO) UnmarshalJSON(data []byte) (err error) {
	// This validates that all required properties are included in the JSON object
	// by unmarshalling the object into a generic map with string keys and checking
	// that every required field exists as a key in the generic map.
	requiredProperties := []string{
		"retention",
	}

	allProperties := make(map[string]interface{})

	err = json.Unmarshal(data, &allProperties)

	if err != nil {
		return err;
	}

	for _, requiredProperty := range(requiredProperties) {
		if _, exists := allProperties[requiredProperty]; !exists {
			return fmt.Errorf("no value given for required property %v", requiredProperty)
		}
	}

	varCreatePrebuildDTO := _CreatePrebuildDTO{}

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&varCreatePrebuildDTO)

	if err != nil {
		return err
	}

	*o = CreatePrebuildDTO(varCreatePrebuildDTO)

	return err
}

type NullableCreatePrebuildDTO struct {
	value *CreatePrebuildDTO
	isSet bool
}

func (v NullableCreatePrebuildDTO) Get() *CreatePrebuildDTO {
	return v.value
}

func (v *NullableCreatePrebuildDTO) Set(val *CreatePrebuildDTO) {
	v.value = val
	v.isSet = true
}

func (v NullableCreatePrebuildDTO) IsSet() bool {
	return v.isSet
}

func (v *NullableCreatePrebuildDTO) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableCreatePrebuildDTO(val *CreatePrebuildDTO) *NullableCreatePrebuildDTO {
	return &NullableCreatePrebuildDTO{value: val, isSet: true}
}

func (v NullableCreatePrebuildDTO) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableCreatePrebuildDTO) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


