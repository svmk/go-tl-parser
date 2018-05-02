package main

import (
	"github.com/svmk/go-tl-parser"
	"encoding/json"
	"strings"
)
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
	result = make([]map[string]interface{}, 0)
	for _, field := range class.Properties {
		item := make(map[string]interface{})
		item["type"] = field.Type
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
			item["returnType"] = class.RootName
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
	for _, iface := range schema.InterfaceInfoes {
		item := map[string]interface{} {
			"fields": []interface{} {},
			"desc": iface.Description,
			"url": nil,
			"extends": "TDObject",
			"type": "object",
		};
		result[iface.Name] = item;
	}
	return json.Marshal(result)
}
