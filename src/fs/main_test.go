package fs

import (
	"testing"

	"os"

	"github.com/stretchr/testify/assert"
)

func createFile(t *testing.T, filepath string) {
	f, err := os.OpenFile(filepath, os.O_WRONLY|os.O_CREATE, 0666)
	assert.NoError(t, err, "create file")
	f.Sync()
	f.Close()
}
