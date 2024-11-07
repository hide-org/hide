package devcontainer

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/docker/docker/pkg/stdcopy"
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

// Docker combines stdout and stderr into a single stream with headers to distinguish between them.
// The StdCopy function demultiplexes this stream back into separate stdout and stderr.
func readOutputFromContainer(ctx context.Context, src io.Reader, stdout, stderr io.Writer) error {
	errCh := make(chan error, 1)

	go func() {
		_, err := stdcopy.StdCopy(stdout, stderr, src)
		errCh <- err
	}()

	select {
	case <-ctx.Done():
		return fmt.Errorf("timeout or cancellation while reading output from container: %w", ctx.Err())
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("error during output copy: %w", err)
		}
		return nil
	}
}
