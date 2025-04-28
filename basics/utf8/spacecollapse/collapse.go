//go:build !solution

package spacecollapse

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

func CollapseSpaces(input string) string {
	var builder strings.Builder
	builder.Grow(len(input))

	var lastCharWasSpace bool

	for i := 0; i < len(input); {
		r, size := utf8.DecodeRuneInString(input[i:])
		if r == utf8.RuneError {
			builder.WriteRune(utf8.RuneError)
			lastCharWasSpace = false
			size = 1
		} else if unicode.IsSpace(r) {
			if !lastCharWasSpace {
				builder.WriteRune(' ')
				lastCharWasSpace = true
			}
		} else {
			builder.WriteRune(r)
			lastCharWasSpace = false
		}
		i += size
	}

	return builder.String()
}
