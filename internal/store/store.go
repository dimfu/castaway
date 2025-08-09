package store

import (
	"fmt"

	"github.com/google/uuid"
)

type Store struct {
	RegistryMap    map[string]*Registry
	UsingChunkSize int64
}

func New(chunksize int64) *Store {
	return &Store{
		RegistryMap:    make(map[string]*Registry),
		UsingChunkSize: chunksize,
	}
}

func (s *Store) generateKey() string {
	return uuid.NewString()
}

func (s *Store) AddToRegistry(secret string, fileInfo *FileInfo) (*Registry, error) {
	key := s.generateKey()
	if _, ok := s.RegistryMap[key]; !ok {
		r := newRegistry(s.UsingChunkSize, key, fileInfo)
		s.RegistryMap[key] = r
		return r, nil
	}
	return nil, fmt.Errorf("Registry with key '%s' already exists", key)
}

func (s *Store) FindRegistry(key string) (*Registry, error) {
	r, ok := s.RegistryMap[key]
	if !ok {
		return r, fmt.Errorf("Could not find registry with key %s", key)
	}
	return r, nil
}

func (s *Store) ClearRegistry(key string) {
	delete(s.RegistryMap, key)
}
