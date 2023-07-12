package extractor

import (
	"regexp"
)

func init() {
	addExtractor("junit5", junitExtract)
}

var pattern = regexp.MustCompile(`(?m)<testcase name="(.*)" classname="(.*)" time="(.*)">\s<failure message="(.*)" type.*?>`)

func junitExtract(r string) []Result {
	result := make([]Result, 0)
	groups := pattern.FindAllStringSubmatch(r, -1)
	for _, g := range groups {
		result = append(result, Result{Name: getOrDefault(g, 1, ""), Class: getOrDefault(g, 2, ""), Time: getOrDefault(g, 3, "0"), Message: getOrDefault(g, 4, "")})
	}
	return result
}

func getOrDefault(s []string, i int, def string) string {
	if i >= len(s) {
		return def
	}
	return s[i]
}
