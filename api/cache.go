package api

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/patrickmn/go-cache"
)

var cachedb *cache.Cache

func cacheStore(key string, value interface{}) error {
	v, err := json.Marshal(value)
	if err != nil {
		return err
	}
	cachedb.SetDefault(key, string(v))
	return nil
}

func cacheStoreCustom(key string, value interface{}, t time.Duration) error {
	v, err := json.Marshal(value)
	if err != nil {
		return err
	}
	cachedb.Set(key, string(v), t)
	return nil
}

func cacheGet(key string, value interface{}) error {
	v, ok := cachedb.Get(key)
	if !ok {
		return errors.New("item not exist")
	}
	return json.Unmarshal([]byte(v.(string)), value)
}
