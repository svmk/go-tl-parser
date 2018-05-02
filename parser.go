package go_tl

import (
	"bufio"
	"errors"
	"io"
	"strings"
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

// ClassInfo holds info of a Class in .tl file
type ClassInfo struct {
	Name        string          `json:"name"`
	Properties  []ClassProperty `json:"properties"`
	Description string          `json:"description"`
	RootName    string          `json:"rootName"`
	IsFunction  bool            `json:"isFunction"`
}

// ClassProperty holds info about properties of a class (or function)
type ClassProperty struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

// InterfaceInfo equals to abstract base classes in .tl file
type InterfaceInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// EnumInfo ...
type EnumInfo struct {
	EnumType string   `json:"enumType"`
	Items    []string `json:"description"`
}

type Schema struct {
	ClassInfoes     []ClassInfo
	InterfaceInfoes []InterfaceInfo
	EnumInfoes      []EnumInfo
}

func replaceKeyWords(input string) string {
	input = strings.Replace(input, "Api", "API", -1)
	input = strings.Replace(input, "Url", "URL", -1)
	input = strings.Replace(input, "Id", "ID", -1)
	input = strings.Replace(input, "Ttl", "TTL", -1)
	input = strings.Replace(input, "Html", "HTML", -1)
	input = strings.Replace(input, "Uri", "URI", -1)
	input = strings.Replace(input, "Ip", "IP", -1)
	input = strings.Replace(input, "Udp", "UDP", -1)

	return input
}

func Parse(f io.Reader) (*Schema, error) {
	var entityDesc string
	paramDescs := make(map[string]string)
	params := make(map[string]string)
	paramsSlice := make([]string, 0)
	classInfoes := make([]ClassInfo, 0)
	interfaceInfoes := make([]InterfaceInfo, 0)
	enumInfoes := make([]EnumInfo, 0)
	hitFunctions := false
	rd := bufio.NewReader(f)
	for {
		line, err := rd.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}

			return nil, errors.New("read file line error: " + err.Error())
		}
		if strings.Contains(line, "---functions---") {
			hitFunctions = true
			continue
		}

		if strings.HasPrefix(line, "//@class ") {
			line = line[len("//@class "):]
			interfaceName := line[:strings.Index(line, " ")]
			line = line[len(interfaceName):]
			line = line[len(" @description "):]
			entityDesc = line[:len(line)-1]
			interfaceInfo := InterfaceInfo{
				Name:        interfaceName,
				Description: strings.Trim(entityDesc, " "),
			}
			interfaceInfoes = append(interfaceInfoes, interfaceInfo)
			enumInfoes = append(enumInfoes, EnumInfo{EnumType: replaceKeyWords(interfaceName)})

		} else if strings.HasPrefix(line, "//@description ") { // Entity description
			line = line[len("//@description "):]
			indexOfFirstSign := strings.Index(line, "@")

			entityDesc = line[:len(line)-1]
			if indexOfFirstSign != -1 {
				entityDesc = line[:indexOfFirstSign]
			}

			if indexOfFirstSign != -1 { // there is some parameter description inline, parse them
				line = line[indexOfFirstSign+1:]
				rd2 := bufio.NewReader(strings.NewReader(line))
				for {
					paramName, _ := rd2.ReadString(' ')
					if paramName == "" {
						break
					}
					paramName = paramName[:len(paramName)-1]
					paramDesc, _ := rd2.ReadString('@')
					if paramDesc == "" {
						paramDesc, _ = rd2.ReadString('\n')

						paramDescs[paramName] = paramDesc[:len(paramDesc)-1]
						break
					}

					paramDescs[paramName] = paramDesc[:len(paramDesc)-1]
				}
			}
		} else if entityDesc != "" && strings.HasPrefix(line, "//@") {
			line = line[len("//@"):]
			rd2 := bufio.NewReader(strings.NewReader(line))
			for {
				paramName, _ := rd2.ReadString(' ')
				if paramName == "" {
					break
				}
				paramName = paramName[:len(paramName)-1]
				paramDesc, _ := rd2.ReadString('@')
				if paramDesc == "" {
					paramDesc, _ = rd2.ReadString('\n')

					paramDescs[paramName] = paramDesc[:len(paramDesc)-1]
					break
				}

				paramDescs[paramName] = paramDesc[:len(paramDesc)-1]
			}

		} else if entityDesc != "" && !strings.HasPrefix(line, "//") && len(line) > 2 {
			entityName := line[:strings.Index(line, " ")]

			line = line[len(entityName)+1:]
			for {
				if strings.Index(line, ":") == -1 {
					break
				}
				paramName := line[:strings.Index(line, ":")]
				line = line[len(paramName)+1:]
				paramType := line[:strings.Index(line, " ")]
				params[paramName] = paramType
				paramsSlice = append(paramsSlice, paramName)
				line = line[len(paramType)+1:]
			}

			rootName := line[len("= ") : len(line)-2]

			var classProps []ClassProperty
			classProps = make([]ClassProperty, 0, 0)

			for _, paramName := range paramsSlice {
				paramType := params[paramName]
				classProp := ClassProperty{
					Name:        paramName,
					Type:        paramType,
					Description: strings.Trim(paramDescs[paramName], " "),
				}
				classProps = append(classProps, classProp)
			}

			classInfoe := ClassInfo{
				Name:        entityName,
				Description: strings.Trim(entityDesc, " "),
				RootName:    rootName,
				Properties:  classProps,
				IsFunction:  hitFunctions,
			}

			classInfoes = append(classInfoes, classInfoe)
			entityDesc = ""
			paramDescs = make(map[string]string)
			params = make(map[string]string)
			paramsSlice = make([]string, 0, 1)
			ok := false
			var enumInfo EnumInfo
			var i int
			for i, enumInfo = range enumInfoes {
				if enumInfo.EnumType == replaceKeyWords(classInfoe.RootName) {
					ok = true
					break
				}
			}
			if ok && !classInfoe.IsFunction {
				enumInfo.Items = append(enumInfo.Items,
					replaceKeyWords(strings.ToUpper(classInfoe.Name[0:1])+classInfoe.Name[1:]))
				enumInfoes[i] = enumInfo
			}
		}
	}
	return &Schema{
		ClassInfoes:     classInfoes,
		InterfaceInfoes: interfaceInfoes,
		EnumInfoes:      enumInfoes,
	}, nil
}
