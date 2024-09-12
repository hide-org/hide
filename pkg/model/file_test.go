package model

import (
	"bufio"
	"reflect"
	"testing"
)

func TestNewLines(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    []Line
		wantErr bool
	}{
		{
			name:    "with new lines",
			content: "hey hey\nhey",
			want:    []Line{{Number: 1, Content: "hey hey"}, {Number: 2, Content: "hey"}},
			wantErr: false,
		},
		{
			name:    "with escaped new lines",
			content: "print('hey\\nhey')\nprint('bye')",
			want:    []Line{{Number: 1, Content: "print('hey\\nhey')"}, {Number: 2, Content: "print('bye')"}},
			wantErr: false,
		},
		{
			name:    "with new lines",
			content: string(make([]byte, bufio.MaxScanTokenSize+1)),
			want:    []Line{{Number: 1, Content: string(make([]byte, bufio.MaxScanTokenSize+1))}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewLines(tt.content)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewLines() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("NewLines() = %v, want %v", got, tt.want)
			}
		})
	}
}
