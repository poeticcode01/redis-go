package hashtable

import (
	"fmt"
	"sync"
)

type RedisDataStore struct {
	RedisMap map[string]interface{}
}

var Cache *RedisDataStore

var once sync.Once

func GetCache() *RedisDataStore {
	once.Do(func() {
		Cache = &RedisDataStore{
			RedisMap: make(map[string]interface{}),
		}
	})
	return Cache
}

func (rds *RedisDataStore) Set(key string, value interface{}) {
	rds.RedisMap[key] = value
}

func (rds *RedisDataStore) Get(key string) (string, error) {
	value, exists := rds.RedisMap[key]

	if !exists {
		return "", fmt.Errorf("key %s not found", key)
	}
	strValue, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("value for key %s is not a string", key)
	}

	return strValue, nil

}
