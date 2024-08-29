package devcontainer

import (
	"bufio"
	"io"

	"github.com/rs/zerolog/log"
)

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
