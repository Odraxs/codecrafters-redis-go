package rdb

import (
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

const rdbFile = "rdb/data.txt"
const rdbExtension = ".rdb"

func GetRDBContent() ([]byte, error) {
	content, err := os.ReadFile(rdbFile)
	if err != nil {
		return nil, err
	}

	result := make([]byte, hex.DecodedLen(len(content)))
	_, err = hex.Decode(result, content)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// Useful functions for latter implementations I think otherwise I remove them at the end of the Extension
func GetPWD() (string, error) {
	result, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return result, nil
}

func FindRDBFile() (string, error) {
	dir, err := GetPWD()
	if err != nil {
		return "", err
	}

	var result string
	err = filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == rdbExtension {
			result = path
			return nil
		}
		return nil
	})

	if err != nil || result == "" {
		return "", fmt.Errorf("failed to find the rdb file")
	}
	return result, nil
}
