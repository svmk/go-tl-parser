package go_tl

import (
	"strings"
)

type ArraySide int

const ArraySideLeft = 0
const ArraySideRight = 1

func MapArrayType(input string, side ArraySide) string {
	if strings.HasPrefix(input, "vector") {
		if side == ArraySideLeft {
			return "[]" + MapArrayType(input[len("vector<"):len(input)-1], side)
		} else if side == ArraySideRight {
			return MapArrayType(input[len("vector<"):len(input)-1], side) + "[]"
		} else {
			panic("Unknown side")
		}
	}
	return input
}
