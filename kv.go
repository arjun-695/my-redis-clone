package main

import (
	"sync"
)

type KV struct {
	mu   sync.RWMutex // "Read- Write mutex"
	data map[string][]byte
}

func NewKV() *KV {// new key-value pair 
	return &KV{
		data: make(map[string][]byte),
	}
}

// SET adds or updates a key-value pair
func (kv *KV) Set(key, val []byte) error {
	kv.mu.Lock()         // Exclusive write lock
	defer kv.mu.Unlock() //executes at the end of the function \
	//unlocks even if there is an error

	//Go doesnt allow keys as byte slices so convert it to string
	kv.data[string(key)] = val
	return nil
}

// Get returns ([]byte and bool) boolean to tell if the key exists or not
func (kv *KV) Get(key []byte) ([]byte, bool) {
	kv.mu.RLock() // multiple goroutines can access the RLock at the same time
	defer kv.mu.RUnlock()

	val, ok := kv.data[string(key)]
	return val, ok
}

func (kv *KV) Delete(key []byte) bool {
	kv.mu.Lock()
	defer kv.mu.Unlock()

	strKey := string(key)
	_, ok := kv.data[strKey]
	delete(kv.data, strKey)

	return ok //key not present
}
