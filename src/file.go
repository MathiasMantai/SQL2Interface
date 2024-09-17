package src

import (
	// "fmt"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
)

// GetFiles retrieves a list of files  within a specified directory.
//
// dir: The path to the directory to be scanned.
//
// Returns:
// - A slice of os.DirEntry objects representing the files and directories found within the specified directory.
// - An error if any occurred during the directory scanning process.
func GetFiles(dir string) ([]os.DirEntry, error) {

	files, readDirError := os.ReadDir(dir)

	if readDirError != nil {
		return nil, readDirError
	}

	return files, nil
}

// GetFileContent retrieves the content of a file within a specified directory.
//
// Parameters:
// - dir: The path to the directory where the file is located.
// - fileName: The name of the file to be read.
//
// Returns:
// - A string containing the content of the file.
// - An error if any occurred during file reading.
func GetFileContent(dir string, fileName string) (string, error) {
	filePath := filepath.Join(dir, fileName)

	fileContent, readError := os.ReadFile(filePath)

	if readError != nil {
		return "", readError
	}

	return string(fileContent), nil
}

// SaveFile writes the provided file content to a specified file within a directory.
//
// Parameters:
// - dir: The path to the directory where the file will be saved.
// - fileName: The name of the file to be created or overwritten.
// - fileContent: The content to be written to the file.
//
// Returns:
// - An error if any occurred during file writing. If the function completes successfully, it returns nil.
func SaveFile(dir string, fileName string, fileContent string) error {

	filePath := filepath.Join(dir, fileName)

	writeError := os.WriteFile(filePath, []byte(fileContent), 0644)

	if writeError != nil {
		return writeError
	}

	return nil
}

// IsDir checks if the specified file path represents a directory.
//
// filePath: The path to the file or directory to be checked.
//
// Returns:
//   - A boolean value indicating whether the specified file path represents a directory.
//     Returns true if the path is a directory, false otherwise.
//   - An error if any occurred during the file or directory check. If the function completes successfully, it returns nil.
func IsDir(filePath string) (bool, error) {

	info, openFileError := os.Stat(filePath)

	if openFileError != nil {
		return false, openFileError
	}

	return info.IsDir(), nil
}

/* YAML */

type Config struct {
	IgnoreFiles     []string                             `yaml:"ignore_files"`
	IgnoreColumns   map[string][]string                  `yaml:"ignore_columns"`
	CombineTables   map[string]TableCombine              `yaml:"combine_tables"`
	Input           string                               `yaml:"input"`
	Output          map[string]map[string]string         `yaml:"output"`
	SingleFile      bool                                 `yaml:"single_file"`
	ArbitraryFields map[string]map[string]ArbitraryField `yaml:"arbitrary_fields"`
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

type ArbitraryField struct {
	Name   string `yaml:"name"`
	TypeGo string `yaml:"type_go"`
	TypeTs string `yaml:"type_ts"`
}

// LoadConfig reads a YAML configuration file and unmarshals its content into a Config struct.
//
// filePath: The path to the YAML configuration file.
//
// Returns:
// - A pointer to a Config struct containing the unmarshalled configuration data.
// - An error if any occurred during file reading or unmarshalling.
func LoadConfig(filePath string) (*Config, error) {

	fileContent, readFileError := os.ReadFile(filePath)

	if readFileError != nil {
		return nil, readFileError
	}

	var conf Config

	unmarshalError := yaml.Unmarshal(fileContent, &conf)

	if unmarshalError != nil {
		return nil, unmarshalError
	}

	return &conf, nil
}
