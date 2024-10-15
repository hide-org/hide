package git

import (
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

type Client interface {
	Checkout(repo Repository, commit string) error
	Clone(url, dst string) (*Repository, error)
}

type ClientImpl struct{}

func NewClient() Client {
	return &ClientImpl{}
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
	_, err := git.PlainClone(dst, false, &git.CloneOptions{
		URL: url,
	})
	if err != nil {
		return nil, err
	}
	return NewLocalRepository(dst)
}
