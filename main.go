package main

import (
	"fmt"
	"path/filepath"
	c "github.com/MathiasMantai/sql2interface/convert"
	f "github.com/MathiasMantai/sql2interface/file"
)

func main() {
	output := "./output"
	dir := "./sql"
	files, err := f.GetFiles(dir)

	// fmt.Println(files)
	if err != nil {
		fmt.Println(err)
	}

	for _, file := range files {
		fileContent, getContentErr := f.GetFileContent(dir, file.Name())

		if getContentErr != nil {
			fmt.Println(getContentErr)
			continue
		}

		parsedData, err := c.ParseSQL(fileContent)

		if err != nil {
			fmt.Println(err)
			continue
		}
		fileName := parsedData.TableName + ".ts"
		parsedInterface := c.CreateInterface(parsedData)
		writeFileError := f.SaveFile(output, fileName, parsedInterface)

		if writeFileError != nil {
			fmt.Println("=> error detected: " + writeFileError.Error())
            continue
		}

		fmt.Printf("=> creating interface %v and saving to %v\n", parsedData.TableName, filepath.Join(output, fileName))
	}
}
