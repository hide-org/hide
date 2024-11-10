package devcontainer

import (
	"bufio"
	"bytes"
	"io"

	"github.com/rs/zerolog/log"
)

// writer for docker logs
type logPipe struct{}

func (lp *logPipe) Write(p []byte) (n int, err error) {
	err = logResponse(bytes.NewReader(p))
	return len(p), err
}

// logs response from Docker API
func logResponse(src io.Reader) error {
	scanner := bufio.NewScanner(src)
	for scanner.Scan() {
		log.Info().Msg(scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}
