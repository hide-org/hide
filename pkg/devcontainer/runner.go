package devcontainer

type Runner interface {
	Run(config Config) (string, error)
	Stop(containerId string) error
	Exec(containerId string, command string) (string, error)
}
