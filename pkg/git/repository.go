package git

import (
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/spf13/afero"
)

type Repository interface {
	Clone(dst string) error
	NeedsUpdate() (bool, error)
	Update() error
	URI() string
}

type RemoteRepository struct {
	uri string
	// TODO: git client
}

func (r *RemoteRepository) URI() string {
	return r.uri
}

func (r *RemoteRepository) Clone(dst string) error {
	// TODO: can it clone the local repo?
	_, err := git.PlainClone(dst, false, &git.CloneOptions{
		URL:      r.URI(),
		Progress: os.Stdout, // TODO: remove or use a logger
	})
	return err
}

func (r *RemoteRepository) NeedsUpdate() (bool, error) {
	return false, nil
}

func (r *RemoteRepository) Update() error {
	// The remote repository is always up to date
	return nil
}

type LocalRepository struct {
	path string
	fs   afero.Fs
	// TODO: git client
}

func (r *LocalRepository) URI() string {
	return r.path
}

func (r *LocalRepository) Clone(dst string) error {
	return copyLocalRepo(r.fs, r.path, dst)
}

func (r *LocalRepository) NeedsUpdate() (bool, error) {
	// TODO: implement using git client
	return true, nil
}

func (r *LocalRepository) Update() error {
    // Open the repository
    repo, err := git.PlainOpen(r.path)
    if err != nil {
        return err
    }

    // Get the worktree
    w, err := repo.Worktree()
    if err != nil {
        return err
    }

    // Pull the latest changes
    err = w.Pull(&git.PullOptions{RemoteName: "origin"})
    if err != nil && err != git.NoErrAlreadyUpToDate {
        return err
    }

    return nil
}

func copyLocalRepo(fs afero.Fs, src, dst string) error {
	// TODO: why is this so hard?
	return afero.Walk(fs, src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		return os.WriteFile(dstPath, data, info.Mode())
	})
}
