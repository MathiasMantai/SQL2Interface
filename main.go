package main

import (
	"github.com/MathiasMantai/sql2interface/src"
)

func main() {
	confDir := "./s2iconfig.yaml"
	sqlParser := src.NewSQL2Interface(confDir)
	sqlParser.Run()
}
