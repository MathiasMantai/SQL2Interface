package file

import (
	"os"
	"path/filepath"
)

func GetFiles(dir string) ([]os.DirEntry, error) {

	files, readDirError := os.ReadDir(dir)

	if readDirError != nil {
		return nil, readDirError
	}

	return files, nil
}

func GetFileContent(dir string, fileName string) (string, error) {
	filePath := filepath.Join(dir, fileName)

	fileContent, readError := os.ReadFile(filePath)

	if readError != nil {
		return "", readError
	}

	return string(fileContent), nil
}

func SaveFile(dir string, fileName string, fileContent string) error {

	filePath := filepath.Join(dir, fileName)

	writeError := os.WriteFile(filePath, []byte(fileContent), 0644)

	if writeError != nil {
		return writeError
	}

	return nil
}

func IsDir(filePath string) (bool, error) {

	info, openFileError := os.Stat(filePath)

	if openFileError != nil {
		return false, openFileError
	}

	return info.IsDir(), nil
}
