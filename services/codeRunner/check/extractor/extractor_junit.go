package extractor

import (
	"code-runner/model"
	"regexp"
)

func init() {
	addExtractor("junit-5", junitExtract)
}

var pattern = regexp.MustCompile(`(?m)<testcase name="(.*)" classname="(.*)" time="(.*)">\s<failure message="(.*)" type.*?>`)

func junitExtract(r string) []*model.Detail {
	result := make([]*model.Detail, 0)
	groups := pattern.FindAllStringSubmatch(r, -1)
	for _, g := range groups {
		result = append(result, &model.Detail{Name: getOrDefault(g, 1, ""), Class: getOrDefault(g, 2, ""), Time: getOrDefault(g, 3, "0"), Message: getOrDefault(g, 4, "")})
	}
	return result
}

func getOrDefault(s []string, i int, def string) string {
	if i >= len(s) {
		return def
	}
	return s[i]
}
