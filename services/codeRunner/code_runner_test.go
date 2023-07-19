package codeRunner

import (
	"fmt"
	"testing"
)

func TestTemplate(t *testing.T) {

	cmd := "java -jar resources/junit-platform-console-standalone-1.9.3.jar --class-path target --select-class {{getSubstringUntil .FileName \".\"}} --reports-dir=./reports --details=tree"

	service := Service{}
	transformedCmd, _ := service.TransformCommand(cmd, TransformParams{FileName: "Test.java"})
	fmt.Println(transformedCmd)
}
