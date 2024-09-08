package hashtable

import (
	"fmt"
	"strconv"
	"sync"
	"time"
)

type ExpiringValue struct {
	Value interface{}
	Px    *time.Time
}

type RedisDataStore struct {
	RedisMap map[string]ExpiringValue
	Mutex    sync.RWMutex
}

var Cache *RedisDataStore

var once sync.Once

func GetCache() *RedisDataStore {
	once.Do(func() {
		Cache = &RedisDataStore{
			RedisMap: make(map[string]ExpiringValue),
		}
	})
	return Cache
}

func (rds *RedisDataStore) Set(key string, value interface{}, px string) error {
	rds.Mutex.Lock()
	defer rds.Mutex.Unlock()

	var expiryTime *time.Time = nil
	fmt.Printf("Key is %s , Value is %s, expiry time is %s\n", key, value, px)

	if px == "" {
		rds.RedisMap[key] = ExpiringValue{
			Value: value,
			Px:    expiryTime,
		}
		return nil

	}

	ms, err := strconv.Atoi(px)
	if err != nil {
		return fmt.Errorf("invalid ttl %s", px)

	}
	ttl := time.Duration(ms) * time.Millisecond
	expiry := time.Now().Add(ttl)
	expiryTime = &expiry

	rds.RedisMap[key] = ExpiringValue{
		Value: value,
		Px:    expiryTime,
	}
	return nil
}

func (rds *RedisDataStore) Get(key string) (string, error) {
	rds.Mutex.Lock()
	defer rds.Mutex.Unlock()

	value, exists := rds.RedisMap[key]

	if !exists {
		return "", fmt.Errorf("key %s not found", key)
	}

	if value.Px != nil && time.Now().After(*value.Px) {
		return "", fmt.Errorf("key %s is expired", key)

	}

	strValue, ok := value.Value.(string)
	if !ok {
		return "", fmt.Errorf("value for key %s is not a string", key)
	}

	return strValue, nil

}
