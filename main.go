package main

import (
	"flag"
	c "github.com/MathiasMantai/sql2interface/convert"
)

func main() {

	output := flag.String("o", "", "source directory to scan for sql files")
	dir := flag.String("i", "", "target directory to save converted interfaces as typescript files")
	flag.Parse()
	sqlParser := c.NewSQL2Interface(*dir, *output)
	sqlParser.Run()
}
