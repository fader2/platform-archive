// Copyright (c) Fader, IP. All Rights Reserved.
// See LICENSE for license information.

package fs

import "strings"
import "fmt"

type Path string

func (p Path) String() string {
	return fmt.Sprintf("Bucket=%s File=%s", p.Bucket(), p.File())
}

func NewPath(path string) Path {
	return Path(strings.Replace(path, "\\", "/", -1))
}

func (p Path) Bucket() string {
	return strings.Split(string(p), "/")[1]
}

func (p Path) File() string {
	return strings.Join(strings.Split(string(p), "/")[2:], "/")
}
