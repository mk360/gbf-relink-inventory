package memory

import (
	"log"
	"strconv"
	"strings"
)

type patternByte struct {
	value    byte
	wildcard bool
}

func parsePattern(pattern string) []patternByte {
	tokens := strings.Fields(pattern)
	if len(tokens) == 0 {
		log.Fatalln("Empty search pattern")
	}
	result := make([]patternByte, len(tokens))
	for i, token := range tokens {
		if token == "??" {
			result[i] = patternByte{wildcard: true}
			continue
		}
		by, _ := strconv.ParseUint(token, 16, 8)
		result[i] = patternByte{value: byte(by)}
	}
	return result
}
