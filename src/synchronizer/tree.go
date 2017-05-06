package synchronizer

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type FSTree struct {
	Ignorer
	items map[string]fsItem
}

func (t *FSTree) SetIgnorer(ignorer Ignorer) {
	t.Ignorer = ignorer
}

func NewFSTree() *FSTree {
	return &FSTree{
		items: make(map[string]fsItem),
	}
}

func NewFSTreeFromFs(root string) (*FSTree, error) {
	t := &FSTree{
		items: make(map[string]fsItem),
	}

	walcFunc := func(path string, info os.FileInfo, err error) error {
		if err != nil || info == nil {
			return nil
		}

		if info.IsDir() && strings.Contains(path, ".git") {
			return filepath.SkipDir
		} else if strings.Contains(path, ".git") {
			return nil
		}

		if strings.Contains(path, ".fader_index") {
			return nil
		}

		path = strings.TrimPrefix(path, filepath.Join(root, "/"))

		if path == "" {
			return nil
		}

		item := fsItem{
			Path:    path,
			Size:    info.Size(),
			ModTime: info.ModTime(),
			IsDir:   info.IsDir(),
		}
		t.items[path] = item
		return nil
	}

	err := filepath.Walk(root, walcFunc)
	return t, err
}

type fsItem struct {
	Path    string
	Size    int64
	ModTime time.Time
	Hash    string
	IsDir   bool
}

func (t FSTree) Encode(writer io.Writer) error {
	enc := json.NewEncoder(writer)
	enc.SetIndent("  ", "  ")
	err := enc.Encode(t.items)
	return err
}

func (t *FSTree) Decode(reader io.Reader) error {
	dec := json.NewDecoder(reader)
	err := dec.Decode(&t.items)
	return err
}

func (t *FSTree) EncodeToFile(path string) error {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, FilesPermission)
	if err != nil {
		return err
	}
	err = t.Encode(f)
	if err != nil {
		return err
	}
	err = f.Close()
	if err != nil {
		return err
	}
	return nil
}

type ops int

const (
	create = iota + 1
	unlink
	rmDir
	mkDir
	change
)

type Op struct {
	Path string
	Op   ops
}

func (prev FSTree) Calculate(current *FSTree) []Op {

	updates := []Op{}

	// check new files
	for path, item := range current.items {

		if strings.HasSuffix(path, SyncInfoFileName) {
			continue
		}

		if prevItem, has := prev.items[path]; !has {
			if prevItem.IsDir {
				updates = append(updates, Op{Path: path, Op: mkDir})
			} else {
				updates = append(updates, Op{Path: path, Op: create})
			}
		} else {
			// check update file
			if !item.IsDir && (!item.ModTime.Equal(prevItem.ModTime) || item.Size != prevItem.Size) {
				updates = append(updates, Op{Path: path, Op: change})
			}
		}
	}

	// check deletes
	for path, prevItem := range prev.items {
		if _, has := current.items[path]; !has {
			if prevItem.IsDir {
				updates = append(updates, Op{Path: path, Op: rmDir})
			} else {
				updates = append(updates, Op{Path: path, Op: unlink})
			}
		}
	}

	return updates
}

func (tree *FSTree) ItemsFromTree(other *FSTree) {
	tree.items = other.items
}
