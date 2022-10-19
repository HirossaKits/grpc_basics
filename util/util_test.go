package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetFileNames(t *testing.T) {
	fs, err := GetFileNames("../public")
	assert.NoError(t, err)
	a := [...]string{"date.txt", "user.txt"}
	assert.ElementsMatch(t, a, fs)
}
