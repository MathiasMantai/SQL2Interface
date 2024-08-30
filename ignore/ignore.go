package ignore

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type S2Ignore struct {
	FileName      string   `json:"file"`
	FilesToIgnore []string `json:"files_to_ignore"`
}

func NewS2Ignore(dir string) *S2Ignore {

	s2i := &S2Ignore{FileName: ".s2iignore"}
	readFileError := s2i.Load(dir)

	if readFileError != nil {
		fmt.Println("=> no .s2iignore was found. skipping...")
	}

	return s2i
}

func (s2 *S2Ignore) Load(dir string) error {
	filePath := filepath.Join(dir, s2.FileName)

	fileContent, readFileError := os.ReadFile(filePath)

	if readFileError != nil {
		s2.FilesToIgnore = []string{}
		return readFileError
	}

	filesToIgnore := strings.Split(string(fileContent), "\r\n")
	s2.FilesToIgnore = filesToIgnore

	return nil
}

func (s2 *S2Ignore) IsIgnored(fileName string) bool {
	for _, ignoreFile := range s2.FilesToIgnore {
		if strings.EqualFold(ignoreFile, fileName) {
			fmt.Printf("=> file %v will be ignored\n", fileName)
			return true
		}
	}

	return false
}
