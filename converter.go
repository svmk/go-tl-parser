package go_tl

import (
	"strings"
)

type ArraySide int

const ArraySideLeft = 0
const ArraySideRight = 1

func ReplaceKeyWords(input string) string {
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

func mapArrayType(input string, side ArraySide) string {
	if strings.HasPrefix(input, "vector") {
		if side == ArraySideLeft {
			return "[]" + mapArrayType(input[len("vector<"):len(input)-1], side)
		} else if side == ArraySideRight {
			return mapArrayType(input[len("vector<"):len(input)-1], side) + "[]"
		} else {
			panic("Unknown side")
		}
	}
	return input
}

func ConvertDataType(input string, side ArraySide, shallReplaceKeywords bool) (string, bool) {
	propType := ""
	isPrimitiveType := true

	input = mapArrayType(input, side)

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

	if shallReplaceKeywords {
		propType = ReplaceKeyWords(propType)
	}

	return propType, isPrimitiveType
}
