package controller

// From https://github.com/ryanuber/go-glob/blob/256dc444b735e061061cf46c809487313d5b0065/glob.go

import "strings"

const globWildcard = "*"

func globAll(patterns []string, subj string) bool {
	for _, pattern := range patterns {
		if glob(pattern, subj) {
			return true
		}
	}

	return false
}

func glob(pattern, subj string) bool {
	if pattern == "" {
		return subj == pattern
	}

	if pattern == globWildcard {
		return true
	}

	parts := strings.Split(pattern, globWildcard)

	if len(parts) == 1 {
		return subj == pattern
	}

	leadingGlob := strings.HasPrefix(pattern, globWildcard)
	trailingGlob := strings.HasSuffix(pattern, globWildcard)
	end := len(parts) - 1

	for i := 0; i < end; i++ {
		idx := strings.Index(subj, parts[i])

		switch i {
		case 0:
			if !leadingGlob && idx != 0 {
				return false
			}
		default:
			if idx < 0 {
				return false
			}
		}

		subj = subj[idx+len(parts[i]):]
	}

	return trailingGlob || strings.HasSuffix(subj, parts[end])
}
