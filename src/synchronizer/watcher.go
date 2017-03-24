package synchronizer

import (
	"context"
)

type Watcher interface {
	Watch(context.Context, string) error
}
