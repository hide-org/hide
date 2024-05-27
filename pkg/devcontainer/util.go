package devcontainer

import (
	"io"

	"github.com/docker/docker/pkg/stdcopy"
)

func ReadOutputFromContainer(src io.Reader, stdout, stderr io.Writer) error {
	if _, err := stdcopy.StdCopy(stdout, stderr, src); err != nil {
		return err
	}

	return nil
}
