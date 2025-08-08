package monitor

import (
	"os"
	"path/filepath"
)

type Storage interface {
	Load() ([]byte, error)
	Save(data []byte) error
}

type FileStorage struct {
	FilePath string
}

func (fs *FileStorage) Load() ([]byte, error) {
	return os.ReadFile(fs.FilePath)
}

func (fs *FileStorage) Save(data []byte) error {
	if err := os.MkdirAll(filepath.Dir(fs.FilePath), 0755); err != nil {
		return err
	}
	return os.WriteFile(fs.FilePath, data, 0644)
}
