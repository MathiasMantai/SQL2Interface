package main

import (
	"errors"
	"flag"
	c "github.com/MathiasMantai/sql2interface/convert"
	"strings"
)

func main() {
	output := flag.String("o", "", "source directory to scan for sql files")
	dir := flag.String("i", "", "target directory to save converted interfaces as typescript files")
	flag.Parse()

	if strings.TrimSpace(*output) == "" || strings.TrimSpace(*dir) == "" {
		panic(errors.New("=> input and output have to be specified"))
	}

	sqlParser := c.NewSQL2Interface(*dir, *output)
	sqlParser.Run()
}
