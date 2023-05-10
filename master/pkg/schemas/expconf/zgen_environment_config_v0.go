// Code generated by gen.py. DO NOT EDIT.

package expconf

import (
	"github.com/docker/docker/api/types"
	"github.com/santhosh-tekuri/jsonschema/v2"

	"github.com/determined-ai/determined/master/pkg/schemas"
)

func (e EnvironmentConfigV0) Image() EnvironmentImageMapV0 {
	if e.RawImage == nil {
		panic("You must call WithDefaults on EnvironmentConfigV0 before .Image")
	}
	return *e.RawImage
}

func (e *EnvironmentConfigV0) SetImage(val EnvironmentImageMapV0) {
	e.RawImage = &val
}

func (e EnvironmentConfigV0) EnvironmentVariables() EnvironmentVariablesMapV0 {
	if e.RawEnvironmentVariables == nil {
		panic("You must call WithDefaults on EnvironmentConfigV0 before .EnvironmentVariables")
	}
	return *e.RawEnvironmentVariables
}

func (e *EnvironmentConfigV0) SetEnvironmentVariables(val EnvironmentVariablesMapV0) {
	e.RawEnvironmentVariables = &val
}

func (e EnvironmentConfigV0) ProxyPorts() ProxyPortsConfigV0 {
	if e.RawProxyPorts == nil {
		panic("You must call WithDefaults on EnvironmentConfigV0 before .ProxyPorts")
	}
	return *e.RawProxyPorts
}

func (e *EnvironmentConfigV0) SetProxyPorts(val ProxyPortsConfigV0) {
	e.RawProxyPorts = &val
}

func (e EnvironmentConfigV0) Ports() map[string]int {
	return e.RawPorts
}

func (e *EnvironmentConfigV0) SetPorts(val map[string]int) {
	e.RawPorts = val
}

func (e EnvironmentConfigV0) RegistryAuth() *types.AuthConfig {
	return e.RawRegistryAuth
}

func (e *EnvironmentConfigV0) SetRegistryAuth(val *types.AuthConfig) {
	e.RawRegistryAuth = val
}

func (e EnvironmentConfigV0) ForcePullImage() bool {
	if e.RawForcePullImage == nil {
		panic("You must call WithDefaults on EnvironmentConfigV0 before .ForcePullImage")
	}
	return *e.RawForcePullImage
}

func (e *EnvironmentConfigV0) SetForcePullImage(val bool) {
	e.RawForcePullImage = &val
}

func (e EnvironmentConfigV0) PodSpec() *PodSpec {
	return e.RawPodSpec
}

func (e *EnvironmentConfigV0) SetPodSpec(val *PodSpec) {
	e.RawPodSpec = val
}

func (e EnvironmentConfigV0) AddCapabilities() []string {
	return e.RawAddCapabilities
}

func (e *EnvironmentConfigV0) SetAddCapabilities(val []string) {
	e.RawAddCapabilities = val
}

func (e EnvironmentConfigV0) DropCapabilities() []string {
	return e.RawDropCapabilities
}

func (e *EnvironmentConfigV0) SetDropCapabilities(val []string) {
	e.RawDropCapabilities = val
}

func (e EnvironmentConfigV0) ParsedSchema() interface{} {
	return schemas.ParsedEnvironmentConfigV0()
}

func (e EnvironmentConfigV0) SanityValidator() *jsonschema.Schema {
	return schemas.GetSanityValidator("http://determined.ai/schemas/expconf/v0/environment.json")
}

func (e EnvironmentConfigV0) CompletenessValidator() *jsonschema.Schema {
	return schemas.GetCompletenessValidator("http://determined.ai/schemas/expconf/v0/environment.json")
}