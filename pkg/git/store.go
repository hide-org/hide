package git

import (
	"errors"

	"github.com/spf13/afero"
)

// TODO: think about concurrency
type RepositoryStore interface {
	// TODO: is there a proper type for uri?
	Exists(uri string) (bool, error)
	Get(uri string) (Repository, error)
	Upsert(repo Repository) (Repository, error)
}

type RemoteRepositoryStore struct {
	// TODO: git client
}

func (s *RemoteRepositoryStore) Get(uri string) (Repository, error) {
	return &RemoteRepository{uri: uri}, nil
}

func (s *RemoteRepositoryStore) Upsert(repo Repository) (Repository, error) {
	return nil, errors.New("remote repository store does not support upserting")
}

func (s *RemoteRepositoryStore) Exists(uri string) (bool, error) {
	// TODO: implement using git client
	return false, nil
}

type FsRepositoryStore struct {
	fs afero.Fs
}

func (s *FsRepositoryStore) Exists(uri string) (bool, error) {
	path := uriToPath(uri)
	exists, err := afero.Exists(s.fs, path)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (s *FsRepositoryStore) Get(uri string) (Repository, error) {
	exists, err := s.Exists(uri)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.New("repository not found")
	}
	return &LocalRepository{path: uriToPath(uri), fs: s.fs}, nil
}

func (s *FsRepositoryStore) Upsert(repo Repository) (Repository, error) {
	// TODO: if exists, overwrite; else create
	return nil, nil
}

func uriToPath(uri string) string {
	// TODO: implement
	return ""
}

type CachedRepositoryStore struct {
	remote RepositoryStore
	cache  RepositoryStore
}

func (s *CachedRepositoryStore) Get(uri string) (Repository, error) {
	// Check if the repository is in the cache
	exists, err := s.cache.Exists(uri)
	if err != nil {
		return nil, err
	}
	if exists {
		// Repository is in cache, but we need to check if it's up-to-date
		repo, err := s.cache.Get(uri)
		if err != nil {
			return nil, err
		}
		needsUpdate, err := repo.NeedsUpdate()
		if err != nil {
			return nil, err
		}
		if needsUpdate {
			err = repo.Update()
			if err != nil {
				return nil, err
			}
		}
		return repo, nil
	}

	// Repository not in cache, fetch it
	repo, err := s.remote.Get(uri)
	if err != nil {
		return nil, err
	}

	// Add to cache
	return s.cache.Upsert(repo)
}
