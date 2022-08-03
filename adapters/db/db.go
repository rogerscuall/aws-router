package db

import (
	"log"

	"github.com/charmbracelet/charm/kv"
)

type Adapter struct {
	db *kv.KV
}

// NewAdapter returns a new Adapter
// It will connect to the DB and Sync to the latest update
func NewAdapter(dbName string) (*Adapter, error) {
	db, err := kv.OpenWithDefaults(dbName)
	if err != nil {
		return nil, err
	}
	db.Sync()
	return &Adapter{db: db}, nil
}

func (da Adapter) CloseDbConnection() {
	err := da.db.Close()
	if err != nil {
		log.Fatalf("Error closing database: %v", err)
	}
}

func (da Adapter) GetVal(key string) ([]byte, error) {
	val, err := da.db.Get([]byte(key))
	if err != nil {
		return []byte{}, err
	}
	return val, nil
}

func (da Adapter) Sync() {
	err := da.db.Sync()
	if err != nil {
		log.Fatalf("Error syncing database: %v", err)
	}
}

func (da Adapter) SetVal(key string, val []byte) error {
	err := da.db.Set([]byte(key), val)
	if err != nil {
		return err
	}
	return nil
}
