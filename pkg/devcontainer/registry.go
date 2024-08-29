package devcontainer

import (
	"encoding/base64"
	"encoding/json"

	"github.com/docker/docker/api/types/registry"
)

type RegistryCredentials interface {
	GetCredentials() (string, error)
}

type DockerHubRegistryCredentials struct {
	username string
	password string
}

func NewDockerHubRegistryCredentials(username, password string) RegistryCredentials {
	return &DockerHubRegistryCredentials{username: username, password: password}
}

// Encodes the credentials as a base64 encoded JSON string
func (c *DockerHubRegistryCredentials) GetCredentials() (string, error) {
	authConfig := registry.AuthConfig{
		Username: c.username,
		Password: c.password,
	}

	encodedJSON, err := json.Marshal(authConfig)
	if err != nil {
		return "", err
	}

	authStr := base64.URLEncoding.EncodeToString(encodedJSON)
	return authStr, nil
}
