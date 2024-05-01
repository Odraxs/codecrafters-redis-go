package storage

import (
	"fmt"
	"sync"
)

type Storage struct {
	db   map[string]string
	lock *sync.RWMutex
}

func NewStorage() *Storage {
	return &Storage{
		db:   make(map[string]string),
		lock: &sync.RWMutex{},
	}
}

func (s *Storage) Set(key string, val string) {
	s.lock.Lock()
	s.db[key] = val
	s.lock.Unlock()
}

func (s *Storage) Get(key string) (string, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	data, exist := s.db[key]
	if !exist {
		return "", fmt.Errorf("key %s doesn't exist", key)
	}

	return data, nil
}
