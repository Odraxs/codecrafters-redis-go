package rdb

import (
	"encoding/hex"
	"os"
)

const rdbFile = "rdb/data.txt"

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
