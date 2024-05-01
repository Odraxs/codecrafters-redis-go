package storage

import (
	"fmt"
	"strconv"
	"sync"
	"time"
)

type dataStorage struct {
	value          string
	expirationTime *time.Time
}

type Storage struct {
	db   map[string]dataStorage
	lock *sync.RWMutex
}

func NewStorage() *Storage {
	return &Storage{
		db:   make(map[string]dataStorage),
		lock: &sync.RWMutex{},
	}
}

func (s *Storage) Set(key string, val string, expirationTime int) {
	s.lock.Lock()
	defer s.lock.Unlock()

	var expiration *time.Time
	if expirationTime == 0 {
		expiration = nil
	}
	if expirationTime > 0 {
		timeInMsUnit := strconv.Itoa(expirationTime) + "ms"
		if d, err := time.ParseDuration(timeInMsUnit); err == nil {
			newTime := time.Now().Add(d)
			expiration = &newTime
		}
	}

	s.db[key] = dataStorage{
		value:          val,
		expirationTime: expiration,
	}
}

func (s *Storage) Get(key string) (string, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	dataStorage, exist := s.db[key]
	if !exist {
		return "", fmt.Errorf("key %s doesn't exist", key)
	}
	if dataStorage.isExpired() {
		delete(s.db, key)
		return "", fmt.Errorf("the key %s expired since %s", key, dataStorage.expirationTime)
	}

	return dataStorage.value, nil
}

func (ds *dataStorage) isExpired() bool {
	if ds.expirationTime == nil {
		return false
	}
	return *ds.expirationTime != time.Time{} &&
		time.Now().After(*ds.expirationTime)
}
