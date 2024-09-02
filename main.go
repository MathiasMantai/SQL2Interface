package main

import (
	"errors"
	"flag"
	c "github.com/MathiasMantai/sql2interface/convert"
	"github.com/MathiasMantai/sql2interface/file"
	"strings"
	"fmt"
)

func main() {
	output := flag.String("o", "", "source directory to scan for sql files")
	dir := flag.String("i", "", "target directory to save converted interfaces as typescript files")
	flag.Parse()

	conf, _ := file.LoadConfig("./s2iconfig.yaml")

	fmt.Println(conf)

	if strings.TrimSpace(*output) == "" || strings.TrimSpace(*dir) == "" {
		panic(errors.New("=> input and output have to be specified"))
	}

	confDir := "./s2iconfig.yaml"

	sqlParser := c.NewSQL2Interface(confDir, *dir, *output)
	sqlParser.Run()
}
