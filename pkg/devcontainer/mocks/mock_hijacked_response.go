package mocks

import (
	"bufio"
	"bytes"
	"io"
	"net"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/stdcopy"
)

func CreateMockHijackedResponse(stdout, stderr string) types.HijackedResponse {
	// Prepare the output in Docker's multiplexed format
	var buf bytes.Buffer
	stdcopy.NewStdWriter(&buf, stdcopy.Stdout).Write([]byte(stdout))
	stdcopy.NewStdWriter(&buf, stdcopy.Stderr).Write([]byte(stderr))

	// Create a reader from the buffer
	reader := bufio.NewReader(&buf)

	return types.HijackedResponse{
		Conn:   &mockConn{},
		Reader: reader,
	}
}

// mockConn implements a minimal version of net.Conn
type mockConn struct{}

func (m *mockConn) Read(b []byte) (n int, err error)   { return 0, io.EOF }
func (m *mockConn) Write(b []byte) (n int, err error)  { return len(b), nil }
func (m *mockConn) Close() error                       { return nil }
func (m *mockConn) LocalAddr() net.Addr                { return nil }
func (m *mockConn) RemoteAddr() net.Addr               { return nil }
func (m *mockConn) SetDeadline(t time.Time) error      { return nil }
func (m *mockConn) SetReadDeadline(t time.Time) error  { return nil }
func (m *mockConn) SetWriteDeadline(t time.Time) error { return nil }
