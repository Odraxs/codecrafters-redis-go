package storage

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/codecrafters-io/redis-starter-go/rdb"
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

func (s *Storage) GetKeys() []string {
	var keys []string
	for key := range s.db {
		keys = append(keys, key)
	}
	return keys
}

func (s *Storage) ReadRDBFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	reader := bufio.NewReader(file)

	err = rdb.CheckMagicNumber(reader)
	if err != nil {
		return err
	}

	err = rdb.SkipMetadata(reader)
	if err != nil {
		return err
	}

	// Read db number
	// FE 00                       # Indicates database selector. db number = 00
	_, err = reader.ReadByte()
	if err != nil {
		return err
	}

	err = s.loadFileContent(reader)
	if err != nil {
		return err
	}
	return nil
}

func (s *Storage) loadFileContent(reader *bufio.Reader) error {
	for {
		opcode, err := reader.ReadByte()
		if err != nil {
			return err
		}

		// End of the RDB file
		if opcode == rdb.END_OPCODE {
			return nil
		}

		if opcode == rdb.OPCODE_SELECTDB {
			err = rdb.ReadSelectDB(reader)
			if err != nil {
				return err
			}
		}

		if opcode == rdb.OPCODE_RESIZEDB {
			err = rdb.ReadResizeDB(reader)
			if err != nil {
				return err
			}
		}

		// Skipping exp time bytes
		if opcode == rdb.OPCODE_EXPIRETIME_MS {
			fmt.Println("SKIPPING MS EXP TIME")
			reader.Discard(9)
		}
		if opcode == rdb.OPCODE_EXPIRETIME {
			fmt.Println("SKIPPING NOT MS EXP TIME")
			reader.Discard(5)
		}

		// Length Encoding
		keyLength, err := rdb.LengthEncodedInt(reader)
		if err != nil {
			return err
		}
		keyBytes := make([]byte, keyLength)
		_, err = reader.Read(keyBytes)
		if err != nil {
			return err
		}

		// Length Encoding
		valueLength, err := rdb.LengthEncodedInt(reader)
		if err != nil {
			return err
		}
		valueBytes := make([]byte, valueLength)
		_, err = reader.Read(valueBytes)
		if err != nil {
			return err
		}

		// TODO: add support por exp time latter
		fmt.Printf("READ KEY: %s, READ VAL: %s\n", string(keyBytes), string(valueBytes))
		s.Set(string(keyBytes), string(valueBytes), 0)
	}
}

func (ds *dataStorage) isExpired() bool {
	if ds.expirationTime == nil {
		return false
	}
	return *ds.expirationTime != time.Time{} &&
		time.Now().After(*ds.expirationTime)
}
