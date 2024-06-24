package util

import (
	"io/fs"
	"os"
	"strings"
)

func LoadEnv(fsys fs.FS, filename string) error {
	content, err := fs.ReadFile(fsys, filename)
	if err != nil {
		return err
	}

	for _, line := range strings.Split(string(content), "\n") {
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			os.Setenv(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
		}
	}

	return nil
}
