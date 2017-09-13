package properties

import "strings"

func (p *Properties) expand(text string, start int) string {
	placeholderStart := strings.Index(text[start:], placeholderPrefix)
	if placeholderStart == -1 {
		return text
	}

	placeholderStart += start
	keyStart := placeholderStart + len(placeholderPrefix)

	keyLen := strings.Index(text[keyStart:], placeholderSuffix)
	if keyLen == -1 {
		return p.expand(text, keyStart)
	}

	keyEnd := keyStart + keyLen
	key := text[keyStart:keyEnd]
	placeholderEnd := keyEnd + len(placeholderSuffix)

	value, valuePresent := p.m[key]
	if !valuePresent {
		return p.expand(text, placeholderEnd)
	}

	placeholderLen := len(placeholderPrefix) + keyLen + len(placeholderSuffix)
	placeholderValueDiff := len(value) - placeholderLen

	return p.expand(text[:placeholderStart]+value+text[placeholderEnd:], placeholderEnd+placeholderValueDiff)
}

// Expand Expand any property placeholders in the given text using the properties from this set
func (p *Properties) Expand(text string) string {
	return p.expand(text, 0)
}
