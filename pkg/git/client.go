package git

import (
	"fmt"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

type Client interface {
	Checkout(repo Repository, commit string) error
	Clone(url, dst string) (*Repository, error)
}

type ClientImpl struct {
	accessToken string
}

func NewClient(accessToken string) Client {
	return &ClientImpl{accessToken: accessToken}
}

func (c *ClientImpl) Checkout(repo Repository, commit string) error {
	if repo.URL.Scheme != "file" {
		return fmt.Errorf("repo url scheme must be 'file'")
	}
	repoPath := repo.URL.Path
	r, err := git.PlainOpen(repoPath)
	if err != nil {
		return err
	}
	w, err := r.Worktree()
	if err != nil {
		return err
	}
	err = w.Checkout(&git.CheckoutOptions{
		Hash: plumbing.NewHash(commit),
	})
	return err
}

func (c *ClientImpl) Clone(url, dst string) (*Repository, error) {
	var opts *git.CloneOptions
	if strings.Contains(url, "ghe.spotify.net") {
		opts = &git.CloneOptions{
			URL: url,
			Auth: &http.BasicAuth{
				Username: "x-access-token",
				Password: c.accessToken,
			},
		}
	} else {
		opts = &git.CloneOptions{
			URL: url,
		}
	}

	_, err := git.PlainClone(dst, false, opts)
	if err != nil {
		return nil, err
	}
	return NewLocalRepository(dst)
}
