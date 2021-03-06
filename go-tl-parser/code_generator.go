package main

import (
	"github.com/svmk/go-tl-parser"
	"fmt"
	"github.com/asaskevich/govalidator"
	"strings"
)

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
func getEnumName(enum string) string {
	return replaceKeyWords(strings.ToUpper(enum[0:1])+enum[1:] + "Enum");
}

func convertDataType(input string) (string, bool) {
	propType := ""
	isPrimitiveType := true

	input = go_tl.MapArrayType(input, go_tl.ArraySideLeft)

	if strings.Contains(input, "string") || strings.Contains(input, "int32") ||
		strings.Contains(input, "int64") {
		propType = strings.Replace(input, "int64", "JSONInt64", 1)

	} else if strings.Contains(input, "Bool") {
		propType = strings.Replace(input, "Bool", "bool", 1)

	} else if strings.Contains(input, "double") {
		propType = strings.Replace(input, "double", "float64", 1)

	} else if strings.Contains(input, "int53") {
		propType = strings.Replace(input, "int53", "int64", 1)

	} else if strings.Contains(input, "bytes") {
		propType = strings.Replace(input, "bytes", "[]byte", 1)

	} else {
		if strings.HasPrefix(input, "[][]") {
			propType = "[][]" + strings.ToUpper(input[len("[][]"):len("[][]")+1]) + input[len("[][]")+1:]
		} else if strings.HasPrefix(input, "[]") {
			propType = "[]" + strings.ToUpper(input[len("[]"):len("[]")+1]) + input[len("[]")+1:]
		} else {
			propType = strings.ToUpper(input[:1]) + input[1:]
			isPrimitiveType = false
		}
	}

	propType = replaceKeyWords(propType)

	return propType, isPrimitiveType
}

func Generate(schema *go_tl.Schema, generatedPackage string) (gnrtdStructs string, gnrtdMethods string) {
	gnrtdStructs = fmt.Sprintf("package %s\n\n", generatedPackage)
	structUnmarshals := ""
	gnrtdStructs += `
	
	import (
		"encoding/json"
		"fmt"
		"strconv"
		"strings"
	)
	
	`
	gnrtdMethods = fmt.Sprintf("package %s\n\n", generatedPackage)
	gnrtdMethods += `
	
	import (
		"encoding/json"
		"fmt"
	)
	
	`

	gnrtdStructs += "type tdCommon struct {\n" +
		"Type string `json:\"@type\"`\n" +
		"Extra string `json:\"@extra\"`\n" +
		"}\n\n"

	gnrtdStructs += `
	// JSONInt64 alias for int64, in order to deal with json big number problem
	type JSONInt64 int64
	
	// MarshalJSON marshals to json
	func (jsonInt *JSONInt64) MarshalJSON() ([]byte, error) {
		intStr := strconv.FormatInt(int64(*jsonInt), 10)
		return []byte(intStr), nil
	}
	
	// UnmarshalJSON unmarshals from json
	func (jsonInt *JSONInt64) UnmarshalJSON(b []byte) error {
		intStr := string(b)
		intStr = strings.Replace(intStr, "\"", "", 2)
		jsonBigInt, err := strconv.ParseInt(intStr, 10, 64)
		if err != nil {
			return err
		}
		*jsonInt = JSONInt64(jsonBigInt)
		return nil
	}
`

	gnrtdStructs += `
		// TdMessage is the interface for all messages send and received to/from tdlib
		type TdMessage interface{
			MessageType() string
		}
`

	for _, enumInfoe := range schema.EnumInfoes {

		gnrtdStructs += fmt.Sprintf(`
				// %s Alias for abstract %s 'Sub-Classes', used as constant-enum here
				type %s string
				`,
			getEnumName(enumInfoe.EnumType),
			getEnumName(enumInfoe.EnumType)[:len(getEnumName(enumInfoe.EnumType))-len("Enum")],
			getEnumName(enumInfoe.EnumType))

		consts := ""
		for _, item := range enumInfoe.Items {
			consts += item + "Type " + getEnumName(enumInfoe.EnumType )+ " = \"" +
				strings.ToLower(item[:1]) + item[1:] + "\"\n"

		}
		gnrtdStructs += fmt.Sprintf(`
				// %s enums
				const (
					%s
				)`, getEnumName(enumInfoe.EnumType)[:len(getEnumName(enumInfoe.EnumType))-len("Enum")], consts)
	}

	for _, interfaceInfo := range schema.InterfaceInfoes {
		interfaceInfo.Name = replaceKeyWords(interfaceInfo.Name)
		typesCases := ""

		gnrtdStructs += fmt.Sprintf("// %s %s \ntype %s interface {\nGet%sEnum() %sEnum\n}\n\n",
			interfaceInfo.Name, interfaceInfo.Description, interfaceInfo.Name, interfaceInfo.Name, interfaceInfo.Name)

		for _, enumInfoe := range schema.EnumInfoes {
			if getEnumName(enumInfoe.EnumType )== interfaceInfo.Name+"Enum" {
				for _, enumItem := range enumInfoe.Items {
					typeName := enumItem
					typeNameCamel := strings.ToLower(typeName[:1]) + typeName[1:]
					typesCases += fmt.Sprintf(`case %s:
						var %s %s
						err := json.Unmarshal(*rawMsg, &%s)
						return &%s, err
						
						`,
						enumItem+"Type", typeNameCamel, typeName,
						typeNameCamel, typeNameCamel)
				}
				break
			}
		}

		structUnmarshals += fmt.Sprintf(`
				func unmarshal%s(rawMsg *json.RawMessage) (%s, error){

					if rawMsg == nil {
						return nil, nil
					}
					var objMap map[string]interface{}
					err := json.Unmarshal(*rawMsg, &objMap)
					if err != nil {
						return nil, err
					}

					switch %sEnum(objMap["@type"].(string)) {
						%s
					default:
						return nil, fmt.Errorf("Error unmarshaling, unknown type:" +  objMap["@type"].(string))
					}
				}
				`, interfaceInfo.Name, interfaceInfo.Name, interfaceInfo.Name,
			typesCases)
	}

	for _, classInfoe := range schema.ClassInfoes {
		if !classInfoe.IsFunction {
			structName := strings.ToUpper(classInfoe.Name[:1]) + classInfoe.Name[1:]
			structName = replaceKeyWords(structName)
			structNameCamel := strings.ToLower(structName[0:1]) + structName[1:]

			hasInterfaceProps := false
			propsStr := ""
			propsStrWithoutInterfaceOnes := ""
			assignStr := fmt.Sprintf("%s.tdCommon = tempObj.tdCommon\n", structNameCamel)
			assignInterfacePropsStr := ""

			// sort.Sort(classInfoe.Properties)
			for i, prop := range classInfoe.Properties {
				propName := govalidator.UnderscoreToCamelCase(prop.Name)
				propName = replaceKeyWords(propName)

				dataType, isPrimitive := convertDataType(prop.Type)
				propsStrItem := ""
				if isPrimitive || checkIsInterface(dataType, schema) {
					propsStrItem += fmt.Sprintf("%s %s `json:\"%s\"` // %s", propName, dataType, prop.Name, prop.Description)
				} else {
					propsStrItem += fmt.Sprintf("%s *%s `json:\"%s\"` // %s", propName, dataType, prop.Name, prop.Description)
				}
				if i < len(classInfoe.Properties)-1 {
					propsStrItem += "\n"
				}

				propsStr += propsStrItem
				if !checkIsInterface(prop.Type, schema) {
					propsStrWithoutInterfaceOnes += propsStrItem
					assignStr += fmt.Sprintf("%s.%s = tempObj.%s\n", structNameCamel, propName, propName)
				} else {
					hasInterfaceProps = true
					assignInterfacePropsStr += fmt.Sprintf(`
						field%s, _  := 	unmarshal%s(objMap["%s"])
						%s.%s = field%s
						`,
						propName, dataType, prop.Name,
						structNameCamel, propName, propName)
				}
			}
			gnrtdStructs += fmt.Sprintf("// %s %s \ntype %s struct {\n"+
				"tdCommon\n"+
				"%s\n"+
				"}\n\n", structName, classInfoe.Description, structName, propsStr)

			gnrtdStructs += fmt.Sprintf("// MessageType return the string telegram-type of %s \nfunc (%s *%s) MessageType() string {\n return \"%s\" }\n\n",
				structName, structNameCamel, structName, classInfoe.Name)

			paramsStr := ""
			paramsDesc := ""
			assingsStr := ""
			for i, param := range classInfoe.Properties {
				propName := govalidator.UnderscoreToCamelCase(param.Name)
				propName = replaceKeyWords(propName)
				dataType, isPrimitive := convertDataType(param.Type)
				paramName := convertToArgumentName(param.Name)

				if isPrimitive || checkIsInterface(dataType, schema) {
					paramsStr += paramName + " " + dataType

				} else { // if is not a primitive, use pointers
					paramsStr += paramName + " *" + dataType
				}

				if i < len(classInfoe.Properties)-1 {
					paramsStr += ", "
				}
				paramsDesc += "\n// @param " + paramName + " " + param.Description

				if isPrimitive || checkIsInterface(dataType, schema) {
					assingsStr += fmt.Sprintf("%s : %s,\n", propName, paramName)
				} else {
					assingsStr += fmt.Sprintf("%s : %s,\n", propName, paramName)
				}
			}

			// Create New... constructors
			gnrtdStructs += fmt.Sprintf(`
				// New%s creates a new %s
				// %s
				func New%s(%s) *%s {
					%sTemp := %s {
						tdCommon: tdCommon {Type: "%s"},
						%s
					}

					return &%sTemp
				}
				`, structName, structName, paramsDesc,
				structName, paramsStr, structName, structNameCamel,
				structName, classInfoe.Name, assingsStr, structNameCamel)

			if hasInterfaceProps {
				gnrtdStructs += fmt.Sprintf(`
					// UnmarshalJSON unmarshal to json
					func (%s *%s) UnmarshalJSON(b []byte) error {
						var objMap map[string]*json.RawMessage
						err := json.Unmarshal(b, &objMap)
						if err != nil {
							return err
						}
						tempObj := struct {
							tdCommon
							%s
						}{}
						err = json.Unmarshal(b, &tempObj)
						if err != nil {
							return err
						}

						%s

						%s	
						
						return nil
					}
					`, structNameCamel, structName, propsStrWithoutInterfaceOnes,
					assignStr, assignInterfacePropsStr)
			}
			if checkIsInterface(classInfoe.RootName, schema) {
				rootName := replaceKeyWords(classInfoe.RootName)
				gnrtdStructs += fmt.Sprintf(`
					// Get%sEnum return the enum type of this object 
					func (%s *%s) Get%sEnum() %sEnum {
						 return %s 
					}

					`,
					rootName,
					strings.ToLower(structName[0:1])+structName[1:],
					structName, rootName, rootName,
					structName+"Type")
			}

		} else {
			methodName := strings.ToUpper(classInfoe.Name[:1]) + classInfoe.Name[1:]
			methodName = replaceKeyWords(methodName)
			returnType := strings.ToUpper(classInfoe.RootName[:1]) + classInfoe.RootName[1:]
			returnType = replaceKeyWords(returnType)
			returnTypeCamel := strings.ToLower(returnType[:1]) + returnType[1:]
			returnIsInterface := checkIsInterface(returnType, schema)

			asterike := "*"
			ampersign := "&"
			if returnIsInterface {
				asterike = ""
				ampersign = ""
			}

			paramsStr := ""
			paramsDesc := ""
			for i, param := range classInfoe.Properties {
				paramName := convertToArgumentName(param.Name)
				dataType, isPrimitive := convertDataType(param.Type)
				if isPrimitive || checkIsInterface(dataType, schema) {
					paramsStr += paramName + " " + dataType

				} else {
					paramsStr += paramName + " *" + dataType
				}

				if i < len(classInfoe.Properties)-1 {
					paramsStr += ", "
				}
				paramsDesc += "\n// @param " + paramName + " " + param.Description
			}

			gnrtdMethods += fmt.Sprintf(`
				// %s %s %s
				func (client *Client) %s(%s) (%s%s, error)`, methodName, classInfoe.Description, paramsDesc, methodName,
				paramsStr, asterike, returnType)

			paramsStr = ""
			for i, param := range classInfoe.Properties {
				paramName := convertToArgumentName(param.Name)

				paramsStr += fmt.Sprintf("\"%s\":   %s,", param.Name, paramName)
				if i < len(classInfoe.Properties)-1 {
					paramsStr += "\n"
				}
			}

			illStr := `fmt.Errorf("error! code: %d msg: %s", result.Data["code"], result.Data["message"])`
			if strings.Contains(paramsStr, returnTypeCamel) {
				returnTypeCamel = returnTypeCamel + "Dummy"
			}
			if returnIsInterface {
				enumType := returnType + "Enum"
				casesStr := ""

				for _, enumInfoe := range schema.EnumInfoes {
					if getEnumName(enumInfoe.EnumType )== enumType {
						for _, item := range enumInfoe.Items {
							casesStr += fmt.Sprintf(`
								case %s:
									var %s %s
									err = json.Unmarshal(result.Raw, &%s)
									return &%s, err
									`, item+"Type", returnTypeCamel, item, returnTypeCamel,
								returnTypeCamel)
						}
						break
					}
				}

				gnrtdMethods += fmt.Sprintf(` {
					result, err := client.SendAndCatch(UpdateData{
						"@type":       "%s",
						%s
					})
	
					if err != nil {
						return nil, err
					}
	
					if result.Data["@type"].(string) == "error" {
						return nil, %s
					}

					switch %s(result.Data["@type"].(string)) {
						%s
					default:
						return nil, fmt.Errorf("Invalid type")
					}
					}
					
					`, classInfoe.Name, paramsStr, illStr,
					enumType, casesStr)

			} else {
				gnrtdMethods += fmt.Sprintf(` {
					result, err := client.SendAndCatch(UpdateData{
						"@type":       "%s",
						%s
					})
	
					if err != nil {
						return nil, err
					}
	
					if result.Data["@type"].(string) == "error" {
						return nil, %s
					}
	
					var %s %s
					err = json.Unmarshal(result.Raw, &%s)
					return %s%s, err
	
					}
					
					`, classInfoe.Name, paramsStr, illStr, returnTypeCamel,
					returnType, returnTypeCamel, ampersign, returnTypeCamel)
			}

		}
	}

	gnrtdStructs += "\n\n" + structUnmarshals
	return gnrtdStructs, gnrtdMethods
}
func convertToArgumentName(input string) string {
	paramName := govalidator.UnderscoreToCamelCase(input)
	paramName = replaceKeyWords(paramName)
	paramName = strings.ToLower(paramName[0:1]) + paramName[1:]
	paramName = strings.Replace(paramName, "type", "typeParam", 1)

	return paramName
}

func checkIsInterface(input string, schema *go_tl.Schema) bool {
	for _, interfaceInfo := range schema.InterfaceInfoes {
		if interfaceInfo.Name == input || replaceKeyWords(interfaceInfo.Name) == input {
			return true
		}
	}

	return false
}
