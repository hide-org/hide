package devcontainer

type DevContainerConfig struct {
	Name                        string                       `json:"name"`
	ForwardPorts                []string                     `json:"forwardPorts"`
	PortsAttributes             map[string]PortAttributes    `json:"portsAttributes"`
	OtherPortsAttributes        PortAttributes               `json:"otherPortsAttributes"`
	ContainerEnv                map[string]string            `json:"containerEnv"`
	RemoteEnv                   map[string]string            `json:"remoteEnv"`
	RemoteUser                  string                       `json:"remoteUser"`
	ContainerUser               string                       `json:"containerUser"`
	UpdateRemoteUserUID         bool                         `json:"updateRemoteUserUID"`
	UserEnvProbe                string                       `json:"userEnvProbe"` // enum: "none", "interactiveShell", "loginShell", or "loginInteractiveShell" (default)
	OverrideCommand             bool                         `json:"overrideCommand"`
	ShutdownAction              string                       `json:"shutdownAction"` // enum: "none", "stopContainer" (default for image or Dockerfile), or "stopCompose" (default for Docker Compose)
	Init                        bool                         `json:"init"`
	Privileged                  bool                         `json:"privileged"`
	CapAdd                      []string                     `json:"capAdd"`
	SecurityOpt                 []string                     `json:"securityOpt"`
	Mounts                      []string                     `json:"mounts"` // can be a string or an object; how to express union types?
	Features                    map[string]map[string]string `json:"features"`
	OverrideFeatureInstallOrder []string                     `json:"overrideFeatureInstallOrder"`
	Customizations              map[string]map[string]any    `json:"customizations"`
}

type PortAttributes struct{}
