package services

import "regexp"

func extractRegex(s, pattern string) string {
	re, err := regexp.Compile(pattern)
	if err != nil {
		Logger.Warn("Invalid regex pattern", "pattern", pattern, "error", err)
		return ""
	}
	m := re.FindStringSubmatch(s)
	if len(m) < 2 {
		return ""
	}
	return m[1]
}
