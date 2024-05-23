package devcontainer

import (
	"encoding/json"
	"fmt"
	"maps"
	"reflect"
	"slices"
	"strconv"
	"strings"
)

type Config struct {
	Path string
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

	// Dockerfile path. Duplicate of BuildProps.Dockerfile.
	Dockerfile string `json:"dockerfile,omitempty"`

	// Docker build context. Duplicate of BuildProps.Context.
	Context string `json:"context,omitempty"`

	// Docker build options
	Build *BuildProps `json:"build,omitempty"`

	// This property accepts a port or array of ports that should be published locally when the container is running. Unlike forwardPorts, your application may need to listen on all interfaces (0.0.0.0) not just localhost for it to be available externally. Defaults to [].
	AppPort AppPort `json:"appPort,omitempty"`

	// Requires workspaceFolder be set as well. Overrides the default local mount point for the workspace when the container is created. Supports the same values as the Docker CLI --mount flag.
	WorkspaceMount *Mount `json:"workspaceMount,omitempty"`

	// Requires workspaceMount be set. Sets the default path that devcontainer.json supporting services / tools should open when connecting to the container. Defaults to the automatic source code mount location.
	WorkspaceFolder string `json:"workspaceFolder,omitempty"`

	// An array of Docker CLI arguments that should be used when running the container. Defaults to [].
	// NOTE: this args are currently ignored because it's too cumbersome to manually parse them into the container.Config or container.HostConfig, there should be a better way to do it
	RunArgs []string `json:"runArgs,omitempty"`
}

func (d *DockerImageProps) Equals(other *DockerImageProps) bool {
	return d.Image == other.Image &&
		d.Build.Equals(other.Build) &&
		slices.Equal(d.AppPort, other.AppPort) &&
		d.WorkspaceMount == other.WorkspaceMount &&
		d.WorkspaceFolder == other.WorkspaceFolder &&
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

	// Path that the Docker build should be run from relative to devcontainer.json. For example, a value of ".." would allow you to reference content in sibling directories. Defaults to ".".
	Context string `json:"context,omitempty"`

	// A set of name-value pairs containing Docker image build arguments that should be passed when building a Dockerfile. Environment and pre-defined variables may be referenced in the values. Defaults to not set. For example: "build": { "args": { "MYARG": "MYVALUE", "MYARGFROMENVVAR": "${localEnv:VARIABLE_NAME}" } }
	Args map[string]*string `json:"args,omitempty"`

	// An array of Docker image build options that should be passed to the build command when building a Dockerfile. Defaults to []. For example: "build": { "options": [ "--add-host=host.docker.internal:host-gateway" ] }
	// NOTE: this options are currently ignored because it's too cumbersome to manually parse them into the types.ImageBuildOptions, there should be a better way to do it
	Options []string `json:"options,omitempty"`

	// A string that specifies a Docker image build target that should be passed when building a Dockerfile. Defaults to not set. For example: "build": { "target": "development" }
	Target string `json:"target,omitempty"`

	// A string or array of strings that specify one or more images to use as caches when building the image. Cached image identifiers are passed to the docker build command with --cache-from.
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
	// A command string or list of command arguments to run on the host machine during initialization, including during container creation and on subsequent starts. The command may run more than once during a given session.
	InitializeCommand LifecycleCommand `json:"initializeCommand,omitempty"`

	// This command is the first of three (along with updateContentCommand and postCreateCommand) that finalizes container setup when a dev container is created. It and subsequent commands execute inside the container immediately after it has started for the first time.
	OnCreateCommand LifecycleCommand `json:"onCreateCommand,omitempty"`

	// This command is the second of three that finalizes container setup when a dev container is created. It executes inside the container after onCreateCommand whenever new content is available in the source tree during the creation process.
	UpdateContentCommand LifecycleCommand `json:"updateContentCommand,omitempty"`

	// This command is the last of three that finalizes container setup when a dev container is created. It happens after updateContentCommand and once the dev container has been assigned to a user for the first time.
	PostCreateCommand LifecycleCommand `json:"postCreateCommand,omitempty"`

	// A command to run each time the container is successfully started.
	// TODO: should we run it inside of container?
	PostStartCommand LifecycleCommand `json:"postStartCommand,omitempty"`

	// A command to run each time a tool has successfully attached to the container.
	// TODO: should we run it inside of container?
	PostAttachCommand LifecycleCommand `json:"postAttachCommand,omitempty"`

	// An enum that specifies the command any tool should wait for before connecting. Defaults to updateContentCommand.
	// TODO: use it
	WaitFor string `json:"waitFor,omitempty"`
}

func (l *LifecycleProps) Equals(other *LifecycleProps) bool {
	return l.InitializeCommand.Equals(&other.InitializeCommand) &&
		l.OnCreateCommand.Equals(&other.OnCreateCommand) &&
		l.UpdateContentCommand.Equals(&other.UpdateContentCommand) &&
		l.PostCreateCommand.Equals(&other.PostCreateCommand) &&
		l.PostStartCommand.Equals(&other.PostStartCommand) &&
		l.PostAttachCommand.Equals(&other.PostAttachCommand) &&
		l.WaitFor == other.WaitFor
}

// Can be useful when provisioning cloud resources
type HostRequirements struct {
	Cpus    int    `json:"cpus,omitempty"`
	Memory  string `json:"memory,omitempty"`
	Storage string `json:"storage,omitempty"`
}

func (h *HostRequirements) Equals(other *HostRequirements) bool {
	return h.Cpus == other.Cpus &&
		h.Memory == other.Memory &&
		h.Storage == other.Storage
}

type GeneralProperties struct {
	// A name for the dev container.
	Name string `json:"name,omitempty"`

	// An array of port numbers or "host:port" values (e.g. [3000, "db:5432"]) that should always be forwarded from inside the primary container to the local machine (including on the web). Defaults to [].
	// TODO: how to use this?
	ForwardPorts []string `json:"forwardPorts,omitempty"`

	// Object that maps a port number, "host:port" value, range, or regular expression to a set of default options.
	PortsAttributes map[string]PortAttributes `json:"portsAttributes,omitempty"`

	// Default options for ports, port ranges, and hosts that aren’t configured using portsAttributes.
	OtherPortsAttributes PortAttributes `json:"otherPortsAttributes,omitempty"`

	// A set of name-value pairs that sets or overrides environment variables for the container. Sets the variable on the Docker container itself, so all processes spawned in the container will have access to it.
	ContainerEnv map[string]string `json:"containerEnv,omitempty"`

	// A set of name-value pairs that sets or overrides environment variables for the devcontainer.json supporting service / tool (or sub-processes like terminals) but not the container as a whole.
	// TODO: how to use it?
	RemoteEnv map[string]string `json:"remoteEnv,omitempty"`

	// Overrides the user that Hide uses to run processes inside the container. Defaults to the user the container as a whole is running as (often root).
	// TODO: use when running commands?
	RemoteUser string `json:"remoteUser,omitempty"`

	// Overrides the user for all operations run as inside the container. Defaults to either root or the last USER instruction in the related Dockerfile used to create the image.
	ContainerUser string `json:"containerUser,omitempty"`

	// On Linux, if containerUser or remoteUser is specified, the user’s UID/GID will be updated to match the local user’s UID/GID to avoid permission problems with bind mounts. Defaults to true.
	// TODO: how to use it?
	UpdateRemoteUserUID bool `json:"updateRemoteUserUID,omitempty"`

	// Indicates the type of shell to use to “probe” for user environment variables: "none", "interactiveShell", "loginShell", or "loginInteractiveShell" (default)
	// NOTE: most likely we don't need this
	UserEnvProbe string `json:"userEnvProbe,omitempty"`

	// Tells devcontainer.json supporting services / tools whether they should run /bin/sh -c "while sleep 1000; do :; done" when starting the container instead of the container’s default command (since the container can shut down if the default command fails). Set to false if the default command must run for the container to function properly. Defaults to true for when using an image or Dockerfile and false when referencing a Docker Compose file.
	// TODO: how to use it?
	OverrideCommand bool `json:"overrideCommand,omitempty"`

	// Indicates whether devcontainer.json supporting tools should stop the containers when the related tool window is closed / shut down. Values are none, stopContainer (default for image or Dockerfile), and stopCompose (default for Docker Compose).
	// TODO: how to use it?
	ShutdownAction string `json:"shutdownAction,omitempty"`

	// Defaults to false. Cross-orchestrator way to indicate whether the tini init process should be used to help deal with zombie processes.
	Init bool `json:"init,omitempty"`

	// Defaults to false. Cross-orchestrator way to cause the container to run in privileged mode (--privileged). Required for things like Docker-in-Docker, but has security implications particularly when running directly on Linux.
	Privileged bool `json:"privileged,omitempty"`

	// Defaults to []. Cross-orchestrator way to add capabilities typically disabled for a container.
	CapAdd []string `json:"capAdd,omitempty"`

	// Defaults to []. Cross-orchestrator way to set container security options.
	SecurityOpt []string `json:"securityOpt,omitempty"`

	// Defaults to unset. Cross-orchestrator way to add additional mounts to a container.
	Mounts []Mount `json:"mounts,omitempty"`

	// An object of Dev Container Feature IDs and related options to be added into your primary container.
	// NOTE: this is not supported yet
	// Features                    map[string]map[string]string `json:"features,omitempty"`

	// By default, Features will attempt to automatically set the order they are installed based on a installsAfter property within each of them. This property allows you to override the Feature install order when needed.
	OverrideFeatureInstallOrder []string `json:"overrideFeatureInstallOrder,omitempty"`

	// Product specific properties, defined in supporting tools like Hide.
	Customizations map[string]map[string]any `json:"customizations,omitempty"`
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
		// maps.Equal(g.Features, other.Features) &&
		slices.Equal(g.OverrideFeatureInstallOrder, other.OverrideFeatureInstallOrder) &&
		customizationsEqual(g.Customizations, other.Customizations)
}

func customizationsEqual(a, b map[string]map[string]any) bool {
	if len(a) != len(b) {
		return false
	}

	for key, value := range a {
		otherValue, ok := b[key]
		if !ok {
			return false
		}

		// this can be slow
		if !reflect.DeepEqual(value, otherValue) {
			return false
		}
	}

	return true
}

type Mount struct {
	//Can be bind, volume, or tmpfs
	Type string `json:"type,omitempty"`

	// May be specified as source or src
	Source string `json:"source,omitempty"`

	// May be specified as destination, dst, or target
	Destination string `json:"destination,omitempty"`
}

func (m *Mount) UnmarshalJSON(data []byte) error {
	var jsonObj interface{}
	err := json.Unmarshal(data, &jsonObj)

	if err != nil {
		return fmt.Errorf("Failed to unmarshal Mount: %w", err)
	}

	switch obj := jsonObj.(type) {
	case string:
		for _, kvPair := range strings.Split(obj, ",") {
			kv := strings.Split(kvPair, "=")
			key := kv[0]
			value := kv[1]

			switch key {
			case "type":
				m.Type = value
			case "source", "src":
				m.Source = value
			case "destination", "dst", "target":
				m.Destination = value
			}
		}

		return nil
	case map[string]interface{}:
		if _type, ok := obj["type"].(string); ok {
			m.Type = _type
		}

		for _, key := range []string{"source", "src"} {
			source, ok := obj[key].(string)
			if ok {
				m.Source = source
				break
			}
		}

		for _, key := range []string{"destination", "dst", "target"} {
			destination, ok := obj[key].(string)
			if ok {
				m.Destination = destination
				break
			}
		}

		return nil

	default:
		return fmt.Errorf("Unsupported type for Mount: %T", jsonObj)
	}
}

type PortAttributes struct {
	Label string `json:"label,omitempty"`

	// enum
	Protocol string `json:"protocol,omitempty"`

	// enum
	OnAutoForward    string `json:"onAutoForward,omitempty"`
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
