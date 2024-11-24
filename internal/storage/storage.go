package storage

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"time"
)

type Tag struct {
	Name string `json:"name"`
}

type Note struct {
	Name      string    `json:"name"`
	Content   string    `json:"content"`
	Tag       Tag       `json:"tag"`
	UpdatedAt time.Time `json:"updated_at"`
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

func (sf *StorageFile) FindNote(name string) (*Note, int) {
	for i := range sf.Notes {
		if sf.Notes[i].Name == name {
			return &sf.Notes[i], i
		}
	}

	return nil, -1
}

func (sf *StorageFile) CreateNote(name string, tag string) error {
	newNote := Note{
		Name:      name,
		Content:   "",
		Tag:       Tag{Name: tag},
		UpdatedAt: time.Now(),
	}

	sf.Notes = append(sf.Notes, newNote)

	return nil
}

func (sf *StorageFile) RecordNote(name string, content string) error {
	note, _ := sf.FindNote(name)

	if note == nil {
		return fmt.Errorf("failed to find the node")
	}

	note.Content = content
	note.UpdatedAt = time.Now()

	return nil
}

func (sf *StorageFile) DeleteNote(name string, tag string) {
	_, indexNote := sf.FindNote(name)

	if indexNote >= 0 {
		sf.Notes = append(sf.Notes[:indexNote], sf.Notes[indexNote+1:]...)
	}
}

func (sf *StorageFile) SortNotesByUpdatedAtDesc() {
	sort.Slice(sf.Notes, func(i, j int) bool {
		return sf.Notes[i].UpdatedAt.After(sf.Notes[j].UpdatedAt)
	})
}

func (sf *StorageFile) Save() error {
	sf.SortNotesByUpdatedAtDesc()

	// Serialize the StorageFile to JSON
	data, err := json.MarshalIndent(*sf, "", "  ")

	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	// Write the JSON data to the file
	file, err := os.Create(sf.Path)

	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	if _, err := file.Write(data); err != nil {
		return fmt.Errorf("failed to write data to file: %w", err)
	}

	return nil
}
