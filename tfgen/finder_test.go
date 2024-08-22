package tfgen

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindConfigFile(t *testing.T) {
	assert := assert.New(t)
	tempDir := t.TempDir()
	println(tempDir)

	os.MkdirAll(tempDir+"/dev/module-a/1/2/3", 0755)
	os.MkdirAll(tempDir+"/dev/module-b/1/2/3", 0755)
	os.MkdirAll(tempDir+"/dev/module-c/1/2/3", 0755)

	os.WriteFile(tempDir+"/.tfgen.yaml", []byte(""), 0644)
	os.WriteFile(tempDir+"/dev/.tfgen.yaml", []byte(""), 0644)
	os.WriteFile(tempDir+"/dev/module-a/1/2/3/.tfgen.yaml", []byte(""), 0644)

	tests := []struct {
		input string
		want  string
	}{
		{input: tempDir + "/dev/", want: tempDir + "/dev/.tfgen.yaml"},
		{input: tempDir + "/dev/module-a/1/2/3", want: tempDir + "/dev/module-a/1/2/3/.tfgen.yaml"},
		{input: tempDir + "/dev/module-a/1/2", want: tempDir + "/dev/.tfgen.yaml"},
		{input: tempDir + "/dev/module-b/1/2/3", want: tempDir + "/dev/.tfgen.yaml"},
	}

	for _, tc := range tests {
		result, _ := findConfigFile(tc.input)
		assert.Equal(tc.want, result, "they should be equal")
	}
}
