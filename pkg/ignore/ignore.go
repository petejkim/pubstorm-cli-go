package ignore

import "strings"

func Parse(content string) []string {
	patterns := strings.Split(content, "\n")

	var filteredPatterns []string
	for _, pattern := range patterns {
		fileteredPattern := strings.Trim(pattern, " \r\n")
		if strings.HasPrefix(fileteredPattern, "#") {
			continue
		}

		if len(fileteredPattern) == 0 {
			continue
		}

		filteredPatterns = append(filteredPatterns, fileteredPattern)
	}

	return filteredPatterns
}
