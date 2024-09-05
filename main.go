package main

import (
	"errors"
	"flag"
	"github.com/MathiasMantai/sql2interface/convert"
	"strings"
)

func main() {
	output := flag.String("o", "", "source directory to scan for sql files")
	dir := flag.String("i", "", "target directory to save converted interfaces as typescript files")
	flag.Parse()

	if strings.TrimSpace(*output) == "" || strings.TrimSpace(*dir) == "" {
		panic(errors.New("=> input and output have to be specified"))
	}

	confDir := "./s2iconfig.yaml"

	sqlParser := convert.NewSQL2Interface(confDir, *dir, *output)
	sqlParser.Run()
}
