package store

import (
	"fmt"
	"sync"

	"github.com/google/uuid"
)

type Store struct {
	sync.RWMutex
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
	s.RLock()
	defer s.RUnlock()
	key := s.generateKey()
	if _, ok := s.RegistryMap[key]; !ok {
		r := newRegistry(s, key, fileInfo)
		s.RegistryMap[key] = r
		return r, nil
	}
	return nil, fmt.Errorf("Registry with key '%s' already exists", key)
}

func (s *Store) FindRegistry(key string) (*Registry, error) {
	s.RLock()
	defer s.RUnlock()
	r, ok := s.RegistryMap[key]
	if !ok {
		return nil, fmt.Errorf("Could not find registry with key %s", key)
	}
	return r, nil
}

func (s *Store) ClearRegistry(key string) {
	s.Lock()
	defer s.Unlock()
	delete(s.RegistryMap, key)
}
