package devcontainer

import (
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"strconv"
)

type Config struct {
	DockerImageProps
	DockerComposeProps
	LifecycleProps
	HostRequirements
	GeneralProperties
}

func (c *Config) Equals(other *Config) bool {
	return c.DockerImageProps.Equals(&other.DockerImageProps) &&
		c.DockerComposeProps.Equals(&other.DockerComposeProps) &&
		c.LifecycleProps.Equals(&other.LifecycleProps) &&
		c.HostRequirements.Equals(&other.HostRequirements) &&
		c.GeneralProperties.Equals(&other.GeneralProperties)
}

type DockerImageProps struct {
	// Required when using an image.
	Image string `json:"image,omitempty"`

	Dockerfile string `json:"dockerfile,omitempty"`

	Context string `json:"context,omitempty"`

	Build BuildProps `json:"build,omitempty"`

	// This property accepts a port or array of ports that should be published locally when the container is running.
	AppPort AppPort `json:"appPort,omitempty"`

	WorkspaceMount string `json:"workspaceMount,omitempty"`

	RunArgs []string `json:"runArgs,omitempty"`
}

func (d *DockerImageProps) Equals(other *DockerImageProps) bool {
	return d.Image == other.Image &&
		d.Build.Equals(&other.Build) &&
		slices.Equal(d.AppPort, other.AppPort) &&
		d.WorkspaceMount == other.WorkspaceMount &&
		slices.Equal(d.RunArgs, other.RunArgs)
}

type AppPort []int

func (a *AppPort) UnmarshalJSON(data []byte) error {
	var jsonObj interface{}
	err := json.Unmarshal(data, &jsonObj)

	if err != nil {
		return fmt.Errorf("Failed to unmarshal AppPort: %w", err)
	}

	switch obj := jsonObj.(type) {
	case string:
		port, err := strconv.Atoi(obj)

		if err != nil {
			return fmt.Errorf("Failed to convert AppPort to int: %w", err)
		}

		*a = []int{port}
		return nil
	case int:
		*a = []int{obj}
		return nil
	case float64:
		*a = []int{int(obj)}
		return nil
	case []interface{}:
		ports := make([]int, 0, len(obj))
		for _, v := range obj {
			switch value := v.(type) {
			case string:
				port, err := strconv.Atoi(value)

				if err != nil {
					return fmt.Errorf("Failed to convert AppPort to int: %w", err)
				}

				ports = append(ports, port)
			case int:
				ports = append(ports, value)
			case float64:
				ports = append(ports, int(value))
			default:
				return fmt.Errorf("Unsupported type for AppPort: %T", value)
			}
		}
		*a = ports
		return nil
	}

	return fmt.Errorf("Unsupported type for AppPort: %T", jsonObj)
}

type BuildProps struct {
	// Required when using a Dockerfile. The location of a Dockerfile that defines the contents of the container.
	// The path is relative to the devcontainer.json file.
	Dockerfile string `json:"dockerfile,omitempty"`

	Context string `json:"context,omitempty"`

	Args map[string]string `json:"args,omitempty"`

	Options []string `json:"options,omitempty"`

	Target string `json:"target,omitempty"`

	CacheFrom StringArray `json:"cacheFrom,omitempty"`
}

func (b *BuildProps) Equals(other *BuildProps) bool {
	return b.Dockerfile == other.Dockerfile &&
		b.Context == other.Context &&
		maps.Equal(b.Args, other.Args) &&
		slices.Equal(b.Options, other.Options) &&
		b.Target == other.Target &&
		slices.Equal(b.CacheFrom, other.CacheFrom)
}

type StringArray []string

func (c *StringArray) UnmarshalJSON(data []byte) error {
	var jsonObj interface{}
	err := json.Unmarshal(data, &jsonObj)

	if err != nil {
		return fmt.Errorf("Failed to unmarshal CacheFrom: %w", err)
	}

	switch obj := jsonObj.(type) {
	case string:
		*c = []string{obj}
		return nil
	case []interface{}:
		strings := make([]string, 0, len(obj))
		for _, v := range obj {
			if value, ok := v.(string); ok {
				strings = append(strings, value)
			} else {
				return fmt.Errorf("Unsupported type for CacheFrom: %T", v)
			}
		}

		*c = strings
		return nil
	}

	return fmt.Errorf("Unsupported type for CacheFrom: %T", jsonObj)
}

type DockerComposeProps struct {
	// Required when using Docker Compose.
	DockerComposeFile StringArray `json:"dockerComposeFile,omitempty"`

	// Required when using Docker Compose.
	Service string `json:"service,omitempty"`

	RunServices []string `json:"runServices,omitempty"`
}

func (d *DockerComposeProps) Equals(other *DockerComposeProps) bool {
	return slices.Equal(d.DockerComposeFile, other.DockerComposeFile) &&
		d.Service == other.Service &&
		slices.Equal(d.RunServices, other.RunServices)
}

type LifecycleProps struct {
	InitializeCommand string `json:"initializeCommand,omitempty"` // string or array or object

	OnCreateCommand string `json:"onCreateCommand,omitempty"` // string or array or object

	UpdateContentCommand string `json:"updateContentCommand,omitempty"` // string or array or object

	PostCreateCommand string `json:"postCreateCommand,omitempty"` // string or array or object

	PostStartCommand string `json:"postStartCommand,omitempty"` // string or array or object

	PostAttachCommand string `json:"postAttachCommand,omitempty"` // string or array or object

	WaitFor string `json:"waitFor,omitempty"` // enum
}

func (l *LifecycleProps) Equals(other *LifecycleProps) bool {
	return l.InitializeCommand == other.InitializeCommand &&
		l.OnCreateCommand == other.OnCreateCommand &&
		l.UpdateContentCommand == other.UpdateContentCommand &&
		l.PostCreateCommand == other.PostCreateCommand &&
		l.PostStartCommand == other.PostStartCommand &&
		l.PostAttachCommand == other.PostAttachCommand &&
		l.WaitFor == other.WaitFor
}

type HostRequirements struct {
	Cpus int `json:"cpus,omitempty"`

	Memory string `json:"memory,omitempty"`

	Storage string `json:"storage,omitempty"`
}

func (h *HostRequirements) Equals(other *HostRequirements) bool {
	return h.Cpus == other.Cpus &&
		h.Memory == other.Memory &&
		h.Storage == other.Storage
}

type GeneralProperties struct {
	Name                 string                    `json:"name,omitempty"`
	ForwardPorts         []string                  `json:"forwardPorts,omitempty"`
	PortsAttributes      map[string]PortAttributes `json:"portsAttributes,omitempty"`
	OtherPortsAttributes PortAttributes            `json:"otherPortsAttributes,omitempty"`
	ContainerEnv         map[string]string         `json:"containerEnv,omitempty"`
	RemoteEnv            map[string]string         `json:"remoteEnv,omitempty"`
	RemoteUser           string                    `json:"remoteUser,omitempty"`
	ContainerUser        string                    `json:"containerUser,omitempty"`
	UpdateRemoteUserUID  bool                      `json:"updateRemoteUserUID,omitempty"`
	UserEnvProbe         string                    `json:"userEnvProbe,omitempty"` // enum: "none", "interactiveShell", "loginShell", or "loginInteractiveShell" (default)
	OverrideCommand      bool                      `json:"overrideCommand,omitempty"`
	ShutdownAction       string                    `json:"shutdownAction,omitempty"` // enum: "none", "stopContainer" (default for image or Dockerfile), or "stopCompose" (default for Docker Compose)
	Init                 bool                      `json:"init,omitempty"`
	Privileged           bool                      `json:"privileged,omitempty"`
	CapAdd               []string                  `json:"capAdd,omitempty"`
	SecurityOpt          []string                  `json:"securityOpt,omitempty"`
	Mounts               []string                  `json:"mounts,omitempty"` // can be a string or an object; how to express union types?
	WorkspaceFolder      string                    `json:"workspaceFolder,omitempty"`
	// Features                    map[string]map[string]string `json:"features,omitempty"`
	OverrideFeatureInstallOrder []string `json:"overrideFeatureInstallOrder,omitempty"`
	// Customizations              map[string]map[string]any    `json:"customizations,omitempty"`
}

func (g *GeneralProperties) Equals(other *GeneralProperties) bool {
	return g.Name == other.Name &&
		slices.Equal(g.ForwardPorts, other.ForwardPorts) &&
		maps.Equal(g.PortsAttributes, other.PortsAttributes) &&
		g.OtherPortsAttributes == other.OtherPortsAttributes &&
		maps.Equal(g.ContainerEnv, other.ContainerEnv) &&
		maps.Equal(g.RemoteEnv, other.RemoteEnv) &&
		g.RemoteUser == other.RemoteUser &&
		g.ContainerUser == other.ContainerUser &&
		g.UpdateRemoteUserUID == other.UpdateRemoteUserUID &&
		g.UserEnvProbe == other.UserEnvProbe &&
		g.OverrideCommand == other.OverrideCommand &&
		g.ShutdownAction == other.ShutdownAction &&
		g.Init == other.Init &&
		g.Privileged == other.Privileged &&
		slices.Equal(g.CapAdd, other.CapAdd) &&
		slices.Equal(g.SecurityOpt, other.SecurityOpt) &&
		slices.Equal(g.Mounts, other.Mounts) &&
		g.WorkspaceFolder == other.WorkspaceFolder &&
		// maps.Equal(g.Features, other.Features) &&
		slices.Equal(g.OverrideFeatureInstallOrder, other.OverrideFeatureInstallOrder) &&
		true
	// maps.Equal(g.Customizations, other.Customizations)
}

type PortAttributes struct {
	Label            string `json:"label,omitempty"`
	Protocol         string `json:"protocol,omitempty"`      // enum
	OnAutoForward    string `json:"onAutoForward,omitempty"` // enum
	RequireLocalPort bool   `json:"requireLocalPort,omitempty"`
	ElevateIfNeeded  bool   `json:"elevateIfNeeded,omitempty"`
}

func (p *PortAttributes) Equals(other *PortAttributes) bool {
	return p.Label == other.Label &&
		p.Protocol == other.Protocol &&
		p.OnAutoForward == other.OnAutoForward &&
		p.RequireLocalPort == other.RequireLocalPort &&
		p.ElevateIfNeeded == other.ElevateIfNeeded
}
