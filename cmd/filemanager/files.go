package filemanager

import (
	"fmt"
	"os"
)

type FileManager struct {
	overwritePasses int
}

func NewFileManager(overwritePasses int) *FileManager {
	if overwritePasses <= 0 {
		overwritePasses = 3
	}
	return &FileManager{
		overwritePasses: overwritePasses,
	}
}

func (fm *FileManager) Delete(path string, deleteType DeleteType) error {
	switch deleteType {
	case DeleteTypeNormal:
		return os.Remove(path)
	case DeleteTypeSecure:
		return fm.secureDelete(path)
	default:
		return fmt.Errorf("invalid delete type: %s", deleteType)
	}
}

func (fm *FileManager) CreateOutput(path string) (*os.File, error) {
	output, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("failed to create output file: %w", err)
	}
	return output, nil
}

func (fm *FileManager) Validate(path string, shouldExist bool) error {
	fileInfo, err := os.Stat(path)

	if shouldExist {
		if os.IsNotExist(err) {
			return fmt.Errorf("file does not exist: %s", path)
		}
		if fileInfo.Size() == 0 {
			return fmt.Errorf("file is empty: %s", path)
		}
	} else {
		if err == nil {
			return fmt.Errorf("file already exists: %s", path)
		}
		if !os.IsNotExist(err) {
			return fmt.Errorf("unexpected error checking file: %w", err)
		}
		return nil
	}

	return nil
}