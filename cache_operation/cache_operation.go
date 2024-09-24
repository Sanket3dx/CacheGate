package cacheoperation

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/dgraph-io/badger"
)

type CacheItem struct {
	Data []byte    `json:"data"`
	TTL  time.Time `json:"ttl"`
}

func NewCacheItem(data []byte, ttl time.Duration) *CacheItem {
	return &CacheItem{
		Data: data,
		TTL:  time.Now().Add(ttl),
	}
}

func (ci *CacheItem) IsExpired() bool {
	return time.Now().After(ci.TTL)
}

func structToBytes(v CacheItem) ([]byte, error) {
	bytes, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func bytesToStruct(data []byte) (CacheItem, error) {
	var item CacheItem
	err := json.Unmarshal(data, &item)
	if err != nil {
		return CacheItem{}, err
	}
	return item, nil
}

func OpenDB(dbName string) *badger.DB {
	opts := badger.DefaultOptions("./" + dbName).WithLogger(nil)
	db, err := badger.Open(opts)
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func Set(db *badger.DB, key string, value CacheItem) error {
	encodedValue, err := structToBytes(value)
	if err != nil {
		return fmt.Errorf("error encoding value: %v", err)
	}
	err = db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), encodedValue)
	})
	if err != nil {
		return fmt.Errorf("error setting value in database: %v", err)
	}
	fmt.Printf("Set: %s\n", key)
	return nil
}

func Get(db *badger.DB, key string) (CacheItem, error) {
	var value CacheItem
	err := db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return fmt.Errorf("key not found: %v", err)
		}
		v, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}
		value, err = bytesToStruct(v)
		if err != nil {
			return fmt.Errorf("error decoding value: %v", err)
		}
		return nil
	})
	if err != nil {
		return value, err
	}
	fmt.Printf("Get: %s\n", key)
	return value, nil
}

func Delete(db *badger.DB, key string) error {
	err := db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(key))
	})
	if err != nil {
		return fmt.Errorf("error deleting key: %v", err)
	}
	fmt.Printf("Deleted key: %s\n", key)
	return nil
}

func CacheGarbageCollector(db *badger.DB) {
	for {
		err := db.View(func(txn *badger.Txn) error {
			opts := badger.DefaultIteratorOptions
			opts.PrefetchValues = true
			it := txn.NewIterator(opts)
			defer it.Close()

			for it.Rewind(); it.Valid(); it.Next() {
				item := it.Item()
				key := item.Key()
				err := item.Value(func(val []byte) error {
					cacheItem, err := bytesToStruct(val)
					if err != nil {
						return err
					}
					if cacheItem.IsExpired() {
						Delete(db, string(key))
					}
					return nil
				})

				if err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			fmt.Printf("Error during iteration: %v\n", err)
		}
		time.Sleep(time.Minute * 1)
	}
}
