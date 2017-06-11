package boltdb

import (
	"os"
	"testing"
	"time"

	"log"

	"github.com/boltdb/bolt"
)

var db *bolt.DB

func TestMain(m *testing.M) {
	var err error
	db, err = bolt.Open(
		"_fortesting.db",
		0600,
		&bolt.Options{
			Timeout: 1 * time.Second,
		},
	)
	if err != nil {
		log.Fatal("open db", err)
	}

	os.Exit(m.Run())
}
