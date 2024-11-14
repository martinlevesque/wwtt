package storage

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type Tag struct {
	Name string `json:"name"`
}

type Note struct {
	Name    string `json:"name"`
	Content string `json:"content"`
	Tag     Tag    `json:"tag"`
}

type StorageFile struct {
	Path  string `json:"-"`
	Notes []Note `json:"notes"`
}

func Init(path string) (*StorageFile, error) {
	// Open the JSON file
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Read the contents of the file
	data, err := io.ReadAll(file)

	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	storageFile := &StorageFile{Path: path}

	if err := json.Unmarshal(data, storageFile); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return storageFile, nil
}
