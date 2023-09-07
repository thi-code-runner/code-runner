package extractor

import (
	"code-runner/model"
	"regexp"
	"strings"
)

func init() {
	addExtractor("junit-5-file", junitExtractFile)
	addExtractor("junit-5-out", junitExtractOut)
}

var patternFile = regexp.MustCompile(`(?m)<testcase name="([\s\S]*?)"[\s\S]*?classname="([\s\S]*?)"[\s\S]*?time="([\s\S]*?)">[\s\S]*?<failure message="([\s\S]*?)"[\s\S]*?type[\s\S]*?>`)
var patternOut = regexp.MustCompile(`(?m)className[\s\S]*?=[\s\S]*?'([\s\S]*?)'[\s\S]*?methodName\s*=\s*'([\s\S]*?)'[\s\S]*?:([\s\S]*?)org\.junit`)

func junitExtractFile(r string) []*model.Detail {
	result := make([]*model.Detail, 0)
	groups := patternFile.FindAllStringSubmatch(r, -1)
	for _, g := range groups {
		result = append(result, &model.Detail{Name: getOrDefault(g, 1, ""), Class: getOrDefault(g, 2, ""), Message: getOrDefault(g, 4, "")})
	}
	return result
}

func junitExtractOut(r string) []*model.Detail {
	result := make([]*model.Detail, 0)
	groups := patternOut.FindAllStringSubmatch(r, -1)
	for _, g := range groups {
		result = append(result, &model.Detail{Name: getOrDefault(g, 1, ""), Class: getOrDefault(g, 2, ""), Message: getOrDefault(g, 3, "")})
	}
	return result
}
func getOrDefault(s []string, i int, def string) string {
	if i >= len(s) {
		return def
	}
	return strings.Trim(s[i], " ")
}
