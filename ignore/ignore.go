package ignore

import (
	"fmt"
	"strings"
	"github.com/MathiasMantai/sql2interface/file"
)

type S2Ignore struct {
	Directory      string   `json:"file"`
	Config  *file.Config	`json:"config"`

}

func NewS2Ignore(confDir string) *S2Ignore {

	s2i := &S2Ignore{Directory: confDir}
	readFileError := s2i.Load()

	if readFileError != nil {
		fmt.Println("=> no .s2iignore was found. skipping...")
	}

	return s2i
}

func (s2 *S2Ignore) Load() error {

	conf, loadConfError := file.LoadConfig(s2.Directory)

	if loadConfError!= nil {
        return loadConfError
    }

	s2.Config = conf

	return nil
}

func (s2 *S2Ignore) IsFileIgnored(fileName string) bool {
	for _, ignoreFile := range s2.Config.IgnoreFiles {
		if strings.EqualFold(ignoreFile, fileName) {
			fmt.Printf("=> file %v will be ignored\n", fileName)
			return true
		}
	}

	return false
}

func (s2 *S2Ignore) IsColumnIgnored(fileName string, columnName string) bool {
	for _, ignoreColumn := range s2.Config.IgnoreColumns[fileName] {
        if strings.EqualFold(ignoreColumn, columnName) {
            fmt.Printf("=> column %v will be ignored\n", columnName)
            return true
        }
    }

    return false
}