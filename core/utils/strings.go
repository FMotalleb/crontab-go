package utils

const (
	escapedCharacter = '\\'
)

func EscapedSplit(s string, sep rune) []string {
	result := make([]string, 0)
	buffer := make([]byte, 0)
	pushBuff := func(r rune) {
		buffer = append(buffer, byte(r))
	}
	escaped := false

	for _, part := range s {
		switch {
		case escaped && part == sep:
			pushBuff(part)
			escaped = false
		case escaped && part != sep:
			pushBuff(escapedCharacter)
			pushBuff(part)
			escaped = false
		case part == escapedCharacter:
			escaped = true
		case part == sep:
			result = append(result, string(buffer))
			buffer = make([]byte, 0)
		default:
			pushBuff(part)
		}
	}
	if len(buffer) > 0 {
		result = append(result, string(buffer))
	}
	if escaped {
		panic("escaped character at the end of string")
	}
	return result
}
