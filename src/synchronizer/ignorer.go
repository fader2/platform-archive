package synchronizer

import (
	"github.com/sabhiram/go-gitignore"
)

type Ignorer interface {
	Ignore(string) bool
}

type GitIgnore struct {
	ign *ignore.GitIgnore
}

func NewGitIgnore(fpath string) (*GitIgnore, error) {
	ign, err := ignore.CompileIgnoreFile(fpath)
	if err != nil {
		return nil, err
	}
	return &GitIgnore{ign: ign}, nil
}

func (i *GitIgnore) Ignore(fpath string) bool {
	return i.ign.MatchesPath(fpath)
}
