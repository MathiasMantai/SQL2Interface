package file

import "os"

func GetFiles(dir string) ([]os.DirEntry, error) {
	
	files, readDirError := os.ReadDir(dir)

	if readDirError!= nil {
        return nil, readDirError
    }

	return files, nil
}