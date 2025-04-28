package reverse

import (
	"strings"
	"unicode/utf8"
)

func Reverse(s string) string {
	var builder strings.Builder
	builder.Grow(len(s))

	i := len(s)
	for i > 0 {
		r, size := utf8.DecodeLastRuneInString(s[:i])
		builder.WriteRune(r)
		if r == utf8.RuneError {
			size = 1
		}
		i -= size
	}

	return builder.String()
}
