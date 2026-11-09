package github

import (
	"regexp"
	"strconv"
	"strings"
)

var taskRefRegex = regexp.MustCompile(`(?i)(?:^|[^a-z0-9_-])([a-z0-9_-]+)-(\d+)\b`)

// extractTaskRefs scans input strings for project slug and task number identifiers.
func extractTaskRefs(texts ...string) []TaskRef {
	seen := make(map[string]bool)
	var refs []TaskRef
	for _, text := range texts {
		if text == "" {
			continue
		}
		matches := taskRefRegex.FindAllStringSubmatch(text, -1)
		for _, match := range matches {
			if len(match) < 3 {
				continue
			}
			slug := strings.ToLower(match[1])
			num, err := strconv.Atoi(match[2])
			if err != nil || num <= 0 {
				continue
			}
			key := slug + "-" + strconv.Itoa(num)
			if seen[key] {
				continue
			}
			seen[key] = true
			refs = append(refs, TaskRef{
				ProjectSlug: slug,
				TaskNumber:  num,
			})
		}
	}
	return refs
}
