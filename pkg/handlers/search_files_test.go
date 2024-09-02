package handlers

import (
	"context"
	"testing"

	"github.com/artmoskvin/hide/pkg/model"
	"github.com/google/go-cmp/cmp"
)

func TestSearch(t *testing.T) {
	// TODO: add tests for context cancellation and errors
	tests := []struct {
		name  string
		ctx   context.Context
		files []*model.File
		query string
		exact bool
		regex bool
		want  []model.File
	}{
		{
			name: "case insensitive search",
			ctx:  context.Background(),
			files: []*model.File{
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
			},
			query: "something",
			want: []model.File{
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
			name: "exact search",
			ctx:  context.Background(),
			files: []*model.File{
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
			},
			exact: true,
			query: "something",
			want: []model.File{
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
			name: "grep search",
			ctx:  context.Background(),
			files: []*model.File{
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
			},
			regex: true,
			query: `^o.*e$`,
			want: []model.File{
				{
					Path: "root/folder2/file2.txt",
					Lines: []model.Line{
						{Number: 0, Content: "only something to see"},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			listFiles := func(ctx context.Context, showHidden bool) ([]*model.File, error) {
				return tt.files, nil
			}

			check, err := getChecker(tt.query, tt.exact, tt.regex)
			if err != nil {
				t.Fatal(err)
			}

			result, err := findInFiles(tt.ctx, listFiles, check)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(tt.want, result); diff != "" {
				t.Errorf("got diff %s", diff)
			}
		})
	}
}
