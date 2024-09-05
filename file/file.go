package file

import (
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"fmt"
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

/* YAML */

type Config struct {
	IgnoreFiles   []string                  `yaml:"ignore_files"`
	IgnoreColumns map[string][]string       `yaml:"ignore_columns"`
	CombineTables map[string]TableCombine `yaml:"combine_tables"`
}

type TableCombine struct {
	Name                string   `yaml:"name"`
	Tables              []string `yaml:"tables"`
	ConvertSingleTables bool     `yaml:"convert_single_tables"`
}

type Field struct {
	Name string `yaml:"name"`
	Type string `yaml:"type"`
}

func LoadConfig(filePath string) (*Config, error) {

	fileContent, readFileError := os.ReadFile(filePath)

	if readFileError != nil {
		panic(readFileError)
	}

	var conf Config

	unmarshalError := yaml.Unmarshal(fileContent, &conf)

	if unmarshalError != nil {
		panic(unmarshalError)
	}

	fmt.Println(conf)

	return &conf, nil
}
