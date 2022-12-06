package csvlogger

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	testOutPath = "testdata"
)

func TestNewFileName(t *testing.T) {
	defer os.RemoveAll(testOutPath)

	// with no files
	os.Mkdir(testOutPath, 0777)

	w := NewWriter(testOutPath, "test")
	name, err := w.nextFileName()
	assert.NoError(t, err)
	assert.Equal(t, testOutPath+"/test0001.csv", name)

	os.Create(name)
	name, err = w.nextFileName()
	assert.NoError(t, err)
	assert.Equal(t, testOutPath+"/test0002.csv", name)

	// with some files
	os.RemoveAll(testOutPath)
	os.Mkdir(testOutPath, 0777)
	os.Create(testOutPath + "/abc.csv")
	os.Create(testOutPath + "/test9x.csv")
	os.Create(testOutPath + "/test0007.csv")

	w = NewWriter(testOutPath, "test")
	name, err = w.nextFileName()
	assert.NoError(t, err)
	assert.Equal(t, testOutPath+"/test0008.csv", name)
}
