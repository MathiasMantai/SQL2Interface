package src

import (
	"fmt"
	"strings"
)

func IsFileIgnored(fileName string, ignoreFiles []string) bool {
	for _, ignoreFile := range ignoreFiles {
		if strings.EqualFold(ignoreFile, fileName) {
			fmt.Printf("  => file %v will be ignored\n", fileName)
			return true
		}
	}

	return false
}

func IsColumnIgnored(fileName string, columnName string, ignoreColumns map[string][]string) bool {
	for _, ignoreColumn := range ignoreColumns[fileName] {
		if strings.EqualFold(ignoreColumn, columnName) {
			fmt.Printf("  => column %v will be ignored\n", columnName)
			return true
		}
	}

	return false
}
