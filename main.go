package main

import (
	"fmt"
	"flag"
	"path/filepath"
	c "github.com/MathiasMantai/sql2interface/convert"
	f "github.com/MathiasMantai/sql2interface/file"
)

func main() {

	output := flag.String("o", "", "source directory to scan for sql files")
	dir := flag.String("i", "", "target directory to save converted interfaces as typescript files")
	flag.Parse()

	files, err := f.GetFiles(*dir)

	if err != nil {
		fmt.Println(err)
	}

	for _, file := range files {
		fileContent, getContentErr := f.GetFileContent(*dir, file.Name())

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
		writeFileError := f.SaveFile(*output, fileName, parsedInterface)

		if writeFileError != nil {
			fmt.Println("=> error detected: " + writeFileError.Error())
            continue
		}

		fmt.Printf("=> creating interface %v and saving to %v\n", parsedData.TableName, filepath.Join(*output, fileName))
	}
}
