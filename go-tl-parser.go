package main

import (
	"./code_generator"
	"./parser"
	"bufio"
	"encoding/json"
	"flag"
	"log"
	"os"
	"os/exec"
)

const (
	generatedPackageDefault = "tdlib"      // the package in which go types and methods will be placed into
	structsFileNameDefault  = "types.go"   // file name for structs
	methodsFileNameDefault  = "methods.go" // file name for methods
	generateDirDefault      = "./tdlib"
	modeDefault             = GenerateModeJSON
)

// GenerateMode indicates whether to create a json file of .tl file, or create .go files
type GenerateMode int

// GenerateModeJSON for generate json, GenerateModeGolang for generate golang
const (
	GenerateModeJSON GenerateMode = iota
	GenerateModeGolang
)

func main() {

	var inputFile string
	var generatedPackage string
	var structsFileName string
	var methodsFileName string
	var generateDir string
	var generateMode GenerateMode

	flag.StringVar(&inputFile, "file", "./schema.tl", ".tl schema file")
	flag.StringVar(&generateDir, "dir", generateDirDefault, "Generate directory")
	flag.StringVar(&generatedPackage, "package", generatedPackageDefault, "Package in which generated files will be a part of")
	flag.StringVar(&structsFileName, "structs-file", structsFileNameDefault, "file name for structs")
	flag.StringVar(&methodsFileName, "methods-file", methodsFileNameDefault, "file name for methods")
	flag.IntVar((*int)(&generateMode), "mode", int(modeDefault), "Generate mod, indicates whether to create json file, or golang files")

	flag.Parse()

	f, err := os.OpenFile(inputFile, os.O_RDONLY, os.ModePerm)
	if err != nil {
		log.Fatalf("open file error: %v", err)
		return
	}
	defer f.Close()
	schema, err := parser.Parse(f)
	if err != nil {
		log.Fatalf("Parse serror: %v", err)
		return
	}

	// Write results in a json file
	os.MkdirAll(generateDir, os.ModePerm)
	jsonFile, err := os.OpenFile(generateDir+"/schema.json", os.O_CREATE|os.O_RDWR, os.ModePerm)
	if err != nil {
		log.Fatalf("error openning file %v", err)
	}
	jsonBytes, _ := json.Marshal(schema.ClassInfoes)
	w := bufio.NewWriter(jsonFile)
	w.Write(jsonBytes)

	os.Remove(generateDir + "/" + structsFileName)
	gnrtdStructs, gnrtdMethods := code_generator.Generate(schema, generatedPackage)
	structsFile, err := os.OpenFile(generateDir+"/"+structsFileName, os.O_CREATE|os.O_RDWR, os.ModePerm)
	defer structsFile.Close()
	if err != nil {
		log.Fatalf("error openning file %v", err)
	}
	wgo := bufio.NewWriter(structsFile)
	wgo.Write([]byte(gnrtdStructs))

	os.Remove(generateDir + "/" + methodsFileName)
	methodsFile, err := os.OpenFile(generateDir+"/"+methodsFileName, os.O_CREATE|os.O_RDWR, os.ModePerm)
	defer methodsFile.Close()

	if err != nil {
		log.Fatalf("error openning file %v", err)
	}
	wgo = bufio.NewWriter(methodsFile)
	wgo.Write([]byte(gnrtdMethods))

	// format files
	cmd := exec.Command("gofmt", "-w", generateDir+"/"+methodsFileName)
	cmd.Run()

	cmd = exec.Command("gofmt", "-w", generateDir+"/"+structsFileName)
	cmd.Run()
}
