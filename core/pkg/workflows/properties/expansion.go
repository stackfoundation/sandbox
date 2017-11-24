package properties

import "strings"
import "bytes"

type expansionError struct {
	placeholders []string
}

func (e *expansionError) Error() string {
	var message bytes.Buffer

	message.WriteString("Could not find value")
	if len(e.placeholders) > 1 {
		message.WriteString("s")
	}

	message.WriteString(" for ")

	for i, placeholder := range e.placeholders {
		message.WriteString("${")
		message.WriteString(placeholder)
		message.WriteString("}")

		if i < len(e.placeholders)-1 {
			message.WriteString(", ")
		}
	}

	return message.String()
}

func (p *Properties) expand(text string, start int, missing []string) (string, error) {
	placeholderStart := strings.Index(text[start:], placeholderPrefix)
	if placeholderStart == -1 {
		var err error
		if len(missing) > 0 {
			err = &expansionError{placeholders: missing}
		}

		return text, err
	}

	placeholderStart += start
	keyStart := placeholderStart + len(placeholderPrefix)

	keyLen := strings.Index(text[keyStart:], placeholderSuffix)
	if keyLen == -1 {
		return p.expand(text, keyStart, missing)
	}

	keyEnd := keyStart + keyLen
	key := text[keyStart:keyEnd]
	placeholderEnd := keyEnd + len(placeholderSuffix)

	value, valuePresent := p.m[key]
	if !valuePresent {
		missing = append(missing, key)
		return p.expand(text, placeholderEnd, missing)
	}

	placeholderLen := len(placeholderPrefix) + keyLen + len(placeholderSuffix)
	placeholderValueDiff := len(value) - placeholderLen

	return p.expand(text[:placeholderStart]+value+text[placeholderEnd:], placeholderEnd+placeholderValueDiff, missing)
}

// Expand Expand any property placeholders in the given text using the properties from this set
func (p *Properties) Expand(text string) (string, error) {
	if len(text) < 1 {
		return text, nil
	}

	return p.expand(text, 0, []string{})
}
