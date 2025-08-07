package store

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

type Store struct {
	RegistryMap map[string]*Registry
}

func New() *Store {
	return &Store{
		RegistryMap: make(map[string]*Registry),
	}
}

func (s *Store) generateKey(input string) string {
	hash := sha256.Sum256([]byte(input))
	return base64.URLEncoding.EncodeToString(hash[:])
}

func (s *Store) AddToRegistry(secret, filename string) (*Registry, error) {
	key := s.generateKey(secret)
	if _, ok := s.RegistryMap[key]; !ok {
		r := newRegistry(key, filename)
		s.RegistryMap[key] = r
		return r, nil
	}
	return nil, fmt.Errorf("Registry with secret '%s' already exists", secret)
}

func (s *Store) FindRegistry(key string) (*Registry, error) {
	r, ok := s.RegistryMap[key]
	if !ok {
		return r, fmt.Errorf("Could not find registry with key %s", key)
	}
	return r, nil
}
