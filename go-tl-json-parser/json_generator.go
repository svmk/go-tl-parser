package main

import (
	"github.com/svmk/go-tl-parser"
	"encoding/json"
	"strings"
	"go-tl-parser"
)

func typeCast(typeName string) string {
	result, _ := go_tl.ConvertDataType(typeName, go_tl.ArraySideRight, false)
	if result == "int32" {
		result = "number"
	}
	return result
}
func lcFirst(str string) string {
	if str == "" {
		return ""
	}
	return string(strings.ToLower(string(str[0]))) + str[1:]
}
func ucFirst(str string) string {
	if str == "" {
		return ""
	}
	return string(strings.ToUpper(string(str[0]))) + str[1:]
}

func appendDot(text string) string {
	if text == "" {
		return text
	}
	if text[:len(text)-1] != "." {
		return text + "."
	}
	return text
}

func fields(class go_tl.ClassInfo) (result []map[string]interface{}) {
	for _, field := range class.Properties {
		item := make(map[string]interface{})
		fieldType := typeCast(field.Type)
		item["type"] = fieldType
		item["name"] = field.Name
		item["desc"] = appendDot(field.Description)
		result = append(result, item)
	}
	return result
}

func Generate(schema *go_tl.Schema) ([]byte, error) {
	result := make(map[string]interface{})
	for _, class := range schema.ClassInfoes {
		item := map[string]interface{}{
			"fields": fields(class),
			"desc":   appendDot(class.Description),
			"url":    nil,
		}
		if class.IsFunction {
			item["extends"] = "TDFunction"
			item["type"] = "function"
			returnType := typeCast(class.RootName)
			item["returnType"] = returnType
		} else {
			if class.RootName != ucFirst(class.Name) {
				item["extends"] = class.RootName
			} else {
				item["extends"] = "TDObject"
			}
			item["type"] = "object"
		}
		result[class.Name] = item
	}
	return json.Marshal(result)
}
