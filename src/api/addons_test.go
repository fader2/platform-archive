package api

import (
	"github.com/BurntSushi/toml"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestCheckCompatibility(t *testing.T) {
	var config = `
[addons]
basic = "0.1"
example = "0.1"
`

	var as addonsSetting
	_, err := toml.DecodeReader(strings.NewReader(config), &as)
	assert.NoError(t, err)

	assert.Equal(t, 2, len(as.Addons))
}
