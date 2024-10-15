package git

import (
	"errors"
	"net/url"
)

type Repository struct {
	URL url.URL
}

func NewRepository(u url.URL) Repository {
	return Repository{URL: u}
}

func NewLocalRepository(path string) (*Repository, error) {
	u, err := url.Parse(path)
	if err != nil {
		return nil, err
	}
	if u.Scheme == "" {
		u.Scheme = "file"
	}
	if u.Scheme != "file" {
		return nil, errors.New("path scheme must be 'file'")
	}
	if u.Host != "" {
		return nil, errors.New("path host must be empty")
	}
	return &Repository{URL: *u}, nil
}
