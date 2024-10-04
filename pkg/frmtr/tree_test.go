package frmtr_test

import (
	"testing"

	"github.com/hide-org/hide/pkg/frmtr"
)

func TestFilesString(t *testing.T) {
	var paths frmtr.Files = []string{
		"home/user/documents/reports/2024/report1.txt",
		"home/user/documents/reports/2024/report2.txt",
		"home/user/documents/reports/2023/report1.txt",
		"home/user/pictures/vacation/2023/photo1.jpg",
		"home/user/pictures/vacation/2023/photo2.jpg",
		"home/user/pictures/vacation/2022/photo1.jpg",
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

	if got := paths.String(); got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}
