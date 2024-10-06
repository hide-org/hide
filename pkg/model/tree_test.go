package model

import (
	"fmt"
	"testing"
)

func TestFilesString(t *testing.T) {
	var paths Files = []*File{
		{Path: "home/user/documents/reports/2024/report1.txt"},
		{Path: "home/user/documents/reports/2024/report2.txt"},
		{Path: "home/user/documents/reports/2023/report1.txt"},
		{Path: "home/user/pictures/vacation/2023/photo1.jpg"},
		{Path: "home/user/pictures/vacation/2023/photo2.jpg"},
		{Path: "home/user/pictures/vacation/2022/photo1.jpg"},
	}

	want := `.
└── home
    └── user
        ├── documents
        │   └── reports
        │       ├── 2023
        │       │   └── report1.txt
        │       └── 2024
        │           ├── report1.txt
        │           └── report2.txt
        └── pictures
            └── vacation
                ├── 2022
                │   └── photo1.jpg
                └── 2023
                    ├── photo1.jpg
                    └── photo2.jpg
`

	// checks that type works as string for formaters
	if got := fmt.Sprintf("%s", paths); got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}
