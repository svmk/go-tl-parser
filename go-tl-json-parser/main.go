package main

import (
	"github.com/svmk/go-tl-parser"
	"flag"
	"fmt"
	"log"
	"os"
)

func main() {
	var inputFile string
	flag.StringVar(&inputFile, "file", "./schema.tl", ".tl schema file")
	flag.Parse()

	f, err := os.OpenFile(inputFile, os.O_RDONLY, os.ModePerm)
	if err != nil {
		log.Fatalf("open file error: %v", err)
		return
	}
	defer f.Close()
	schema, err := go_tl.Parse(f)
	if err != nil {
		log.Fatalf("Parse serror: %v", err)
		return
	}
	result, err := Generate(schema)
	fmt.Println(string(result))
}
