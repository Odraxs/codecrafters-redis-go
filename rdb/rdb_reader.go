package rdb

import (
	"encoding/hex"
	"os"
)

const rdbFile = "data.txt"

func GetRDBContent() ([]byte, error) {
	err := os.Chdir("rdb")
	if err != nil {
		return nil, err
	}

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
