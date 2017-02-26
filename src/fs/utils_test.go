// Copyright (c) Fader, IP. All Rights Reserved.
// See LICENSE for license information.

package fs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPathUnix(t *testing.T) {
	p := NewPath("/Users/u1/work/src/fader2/platform")
	assert.Equal(t, "Users", p.Bucket())
	assert.Equal(t, "u1/work/src/fader2/platform", p.File())
}

func TestPathWin(t *testing.T) {
	p := NewPath("C:\\Users\\u1\\AppData\\Local\\Temp\\1\\fader928146193\\fader.testfile")
	assert.Equal(t, "Users", p.Bucket())
	assert.Equal(t, "u1/AppData/Local/Temp/1/fader928146193/fader.testfile", p.File())
}
