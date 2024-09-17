package src

import (
	"fmt"
	"strings"
)

// IsFileIgnored checks if a given file name is present in a list of ignored files.//+
// It returns true if the file is ignored, and false otherwise.//+
// //+
// Parameters://+
// - fileName: The name of the file to check.//+
// - ignoreFiles: A slice of strings containing the names of the files to be ignored.//+
// //+
// Return://+
// - A boolean value indicating whether the file is ignored. If true, the file is ignored.//+
func IsFileIgnored(fileName string, ignoreFiles []string) bool {
	for _, ignoreFile := range ignoreFiles {
		if strings.EqualFold(ignoreFile, fileName) {
			fmt.Printf("  => file %v will be ignored\n", fileName)
			return true
		}
	}

	return false
}

// IsColumnIgnored checks if a given column name is present in a list of ignored columns for a specific file.//+
// It returns true if the column is ignored, and false otherwise.//+
// //+
// Parameters://+
// - fileName: The name of the file to check.//+
// - columnName: The name of the column to check.//+
// - ignoreColumns: A map where the keys are file names and the values are slices of strings containing the names of the columns to be ignored.//+
// //+
// Return://+
// - A boolean value indicating whether the column is ignored. If true, the column is ignored.//+
func IsColumnIgnored(fileName string, columnName string, ignoreColumns map[string][]string) bool {
	for _, ignoreColumn := range ignoreColumns[fileName] {
		if strings.EqualFold(ignoreColumn, columnName) {
			fmt.Printf("  => column %v will be ignored\n", columnName)
			return true
		}
	}

	return false
}
