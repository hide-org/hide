package config

import "fmt"

var (
	VersionBranch = "branch"
	VersionCommit = "commit"
	VersionDate   = "0000-00-00 00:00:00+00:00"
)

func Version() string {
	return fmt.Sprintf("%s/%s (%s)", VersionBranch, VersionCommit, VersionDate)
}
