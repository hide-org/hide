package handlers_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hide-org/hide/pkg/files"
	mockfiles "github.com/hide-org/hide/pkg/files/mocks"
	"github.com/hide-org/hide/pkg/handlers"
	"github.com/hide-org/hide/pkg/model"
	"github.com/hide-org/hide/pkg/project/mocks"
)

func TestSearchFileHandler(t *testing.T) {
	modelFiles := []*model.File{
		{
			Path: "root/folder1/file1.txt",
			Lines: []model.Line{
				{Number: 0, Content: "something"},
				{Number: 1, Content: "here is nothing to see"},
			},
		},
		{
			Path: "root/folder2/file2.txt",
			Lines: []model.Line{
				{Number: 0, Content: "only something to see"},
				{Number: 1, Content: "Something"},
			},
		},
	}

	// run tests
	tests := []struct {
		name           string
		ctx            context.Context
		method         string
		target         string
		wantStatusCode int
		wantBody       []model.File
		wantFilter     files.PatternFilter
		wantErr        bool
	}{
		{
			name:           "ok case insensitive search",
			ctx:            context.Background(),
			method:         http.MethodGet,
			target:         "/projects/p1/search?type=content&query=something",
			wantStatusCode: http.StatusOK,
			wantBody: []model.File{
				{
					Path: "root/folder1/file1.txt",
					Lines: []model.Line{
						{Number: 0, Content: "something"},
					},
				},
				{
					Path: "root/folder2/file2.txt",
					Lines: []model.Line{
						{Number: 0, Content: "only something to see"},
						{Number: 1, Content: "Something"},
					},
				},
			},
		},
		{
			name:   "ok case insensitive search with pattern filter",
			ctx:    context.Background(),
			method: http.MethodGet,
			target: "/projects/p1/search?type=content&query=something&include=*.json&include=*.txt&exclude=node",
			wantFilter: files.PatternFilter{
				Include: []string{"*.json", "*.txt"},
				Exclude: []string{"node"},
			},
			wantStatusCode: http.StatusOK,
			wantBody: []model.File{
				{
					Path: "root/folder1/file1.txt",
					Lines: []model.Line{
						{Number: 0, Content: "something"},
					},
				},
				{
					Path: "root/folder2/file2.txt",
					Lines: []model.Line{
						{Number: 0, Content: "only something to see"},
						{Number: 1, Content: "Something"},
					},
				},
			},
		},
		{
			name:           "ok exact search",
			ctx:            context.Background(),
			method:         http.MethodGet,
			target:         "/projects/p1/search?type=content&query=something&exact",
			wantStatusCode: http.StatusOK,
			wantBody: []model.File{
				{
					Path: "root/folder1/file1.txt",
					Lines: []model.Line{
						{Number: 0, Content: "something"},
					},
				},
				{
					Path: "root/folder2/file2.txt",
					Lines: []model.Line{
						{Number: 0, Content: "only something to see"},
					},
				},
			},
		},
		{
			name:           "ok regex search",
			ctx:            context.Background(),
			method:         http.MethodGet,
			target:         "/projects/p1/search?type=content&query=^o.*e$&regex",
			wantStatusCode: http.StatusOK,
			wantBody: []model.File{
				{
					Path: "root/folder2/file2.txt",
					Lines: []model.Line{
						{Number: 0, Content: "only something to see"},
					},
				},
			},
		},
		{
			name: "context cancelled",
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()

				return ctx
			}(),
			method:         http.MethodGet,
			target:         "/projects/p1/search?type=content&query=^o.*e$&regex",
			wantStatusCode: http.StatusInternalServerError, // NOTE: I think we should return 204 No Content
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := handlers.SearchFilesHandler{
				ProjectManager: &mocks.MockProjectManager{
					ListFilesFunc: func(ctx context.Context, projectId string, opts ...files.ListFileOption) (model.Files, error) {
						diff := mockfiles.DiffListFilesOpts(files.ListFilesOptions{
							WithContent: true,
							Filter:      tt.wantFilter,
						}, opts...)

						if diff != "" {
							return nil, fmt.Errorf("filter does not match, diff %s", diff)
						}

						return modelFiles, nil
					},
				},
			}
			r := handlers.NewRouter().WithSearchFileHandler(h).Build()

			req := httptest.NewRequest(tt.method, tt.target, nil).WithContext(tt.ctx)
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			res := rec.Result()
			if tt.wantStatusCode != res.StatusCode {
				t.Fatalf("got status code %v want %v", res.StatusCode, tt.wantStatusCode)
			}
			defer res.Body.Close()

			if tt.wantErr {
				return
			}

			body, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatal(err)
			}

			var out []model.File
			if err := json.Unmarshal(body, &out); err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(tt.wantBody, out); diff != "" {
				t.Errorf("got diff %s", diff)
			}
		})
	}
}
