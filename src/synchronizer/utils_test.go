package synchronizer

import (
	"github.com/boltdb/bolt"
	"github.com/stretchr/testify/assert"
	"interfaces"
	bdb "store/boltdb"
	"testing"
	"time"
)

func TestUtils(t *testing.T) {
	db, err := bolt.Open("/home/god/go/src/github.com/inpime/fader/_app.db", FilesPermission, &bolt.Options{
		Timeout: 1 * time.Second,
	})
	defer db.Close()
	assert.Nil(t, err)
	fm := bdb.NewFileManager(db)

	f, e := fm.FindFileByName("ex1", "noop", interfaces.FullFile)
	assert.Nil(t, f)
	t.Log(e)

	f, e = fm.FindFileByName("not_existing_bucket", "noop", interfaces.FullFile)
	assert.Nil(t, f)
	t.Log(e)
}
