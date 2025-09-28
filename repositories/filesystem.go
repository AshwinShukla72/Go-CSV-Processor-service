package repositories

import (
	"fmt"
	"os"
	"path/filepath"
)

type (
	Store interface {
		SaveOriginal(id string, data []byte) error
		SaveProcessed(id string, data []byte) error
		LoadOriginal(id string) ([]byte, error)
		LoadProcessed(id string) ([]byte, error)
	}
	fileStore struct {
		uploadDir    string
		processedDir string
	}
)

func NewStore(base string) Store {
	upload := filepath.Join(base, "upload")
	processed := filepath.Join(base, "processed")

	// Ensuring both directories exist
	os.MkdirAll(upload, 0755)
	os.MkdirAll(processed, 0755)

	return &fileStore{
		uploadDir:    upload,
		processedDir: processed,
	}
}

func (fs *fileStore) SaveOriginal(id string, data []byte) error {
	fmt.Println(fs.uploadDir)
	uploadPath := filepath.Join(fs.uploadDir, id+".input")
	return os.WriteFile(uploadPath, data, 0644)
}

func (fs *fileStore) SaveProcessed(id string, data []byte) error {
	processedPath := filepath.Join(fs.processedDir, id+".output")
	return os.WriteFile(processedPath, data, 0644)
}

func (fs *fileStore) LoadOriginal(id string) ([]byte, error) {
	uploadPath := filepath.Join(fs.uploadDir, id+".input")
	return os.ReadFile(uploadPath)
}

func (fs *fileStore) LoadProcessed(id string) ([]byte, error) {
	processedPath := filepath.Join(fs.processedDir, id+".output")
	return os.ReadFile(processedPath)
}
