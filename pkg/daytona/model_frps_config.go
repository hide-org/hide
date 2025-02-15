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

// checks if the FRPSConfig type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &FRPSConfig{}

// FRPSConfig struct for FRPSConfig
type FRPSConfig struct {
	Domain string `json:"domain"`
	Port int32 `json:"port"`
	Protocol string `json:"protocol"`
}

type _FRPSConfig FRPSConfig

// NewFRPSConfig instantiates a new FRPSConfig object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewFRPSConfig(domain string, port int32, protocol string) *FRPSConfig {
	this := FRPSConfig{}
	this.Domain = domain
	this.Port = port
	this.Protocol = protocol
	return &this
}

// NewFRPSConfigWithDefaults instantiates a new FRPSConfig object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewFRPSConfigWithDefaults() *FRPSConfig {
	this := FRPSConfig{}
	return &this
}

// GetDomain returns the Domain field value
func (o *FRPSConfig) GetDomain() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Domain
}

// GetDomainOk returns a tuple with the Domain field value
// and a boolean to check if the value has been set.
func (o *FRPSConfig) GetDomainOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Domain, true
}

// SetDomain sets field value
func (o *FRPSConfig) SetDomain(v string) {
	o.Domain = v
}

// GetPort returns the Port field value
func (o *FRPSConfig) GetPort() int32 {
	if o == nil {
		var ret int32
		return ret
	}

	return o.Port
}

// GetPortOk returns a tuple with the Port field value
// and a boolean to check if the value has been set.
func (o *FRPSConfig) GetPortOk() (*int32, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Port, true
}

// SetPort sets field value
func (o *FRPSConfig) SetPort(v int32) {
	o.Port = v
}

// GetProtocol returns the Protocol field value
func (o *FRPSConfig) GetProtocol() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Protocol
}

// GetProtocolOk returns a tuple with the Protocol field value
// and a boolean to check if the value has been set.
func (o *FRPSConfig) GetProtocolOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Protocol, true
}

// SetProtocol sets field value
func (o *FRPSConfig) SetProtocol(v string) {
	o.Protocol = v
}

func (o FRPSConfig) MarshalJSON() ([]byte, error) {
	toSerialize,err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o FRPSConfig) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["domain"] = o.Domain
	toSerialize["port"] = o.Port
	toSerialize["protocol"] = o.Protocol
	return toSerialize, nil
}

func (o *FRPSConfig) UnmarshalJSON(data []byte) (err error) {
	// This validates that all required properties are included in the JSON object
	// by unmarshalling the object into a generic map with string keys and checking
	// that every required field exists as a key in the generic map.
	requiredProperties := []string{
		"domain",
		"port",
		"protocol",
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

	varFRPSConfig := _FRPSConfig{}

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&varFRPSConfig)

	if err != nil {
		return err
	}

	*o = FRPSConfig(varFRPSConfig)

	return err
}

type NullableFRPSConfig struct {
	value *FRPSConfig
	isSet bool
}

func (v NullableFRPSConfig) Get() *FRPSConfig {
	return v.value
}

func (v *NullableFRPSConfig) Set(val *FRPSConfig) {
	v.value = val
	v.isSet = true
}

func (v NullableFRPSConfig) IsSet() bool {
	return v.isSet
}

func (v *NullableFRPSConfig) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableFRPSConfig(val *FRPSConfig) *NullableFRPSConfig {
	return &NullableFRPSConfig{value: val, isSet: true}
}

func (v NullableFRPSConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableFRPSConfig) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


