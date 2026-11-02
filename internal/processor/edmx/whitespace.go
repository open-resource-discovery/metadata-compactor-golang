package edmx

import "strings"

// JavaScript trim includes BOM and excludes NEL; Go's TrimSpace differs.
func isWhitespaceOnly(value string) bool {
	return strings.TrimFunc(value, func(r rune) bool {
		return r == '\t' || r == '\n' || r == '\v' || r == '\f' || r == '\r' ||
			r == ' ' || r == '\u00A0' || r == '\u1680' ||
			(r >= '\u2000' && r <= '\u200A') || r == '\u2028' || r == '\u2029' ||
			r == '\u202F' || r == '\u205F' || r == '\u3000' || r == '\uFEFF'
	}) == ""
}
