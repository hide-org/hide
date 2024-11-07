package config

import "fmt"

var (
	Major = 0
	Minor = 4
	Patch = 1
)

func Version() string {
	return fmt.Sprintf("v%d.%d.%d", Major, Minor, Patch)
}
