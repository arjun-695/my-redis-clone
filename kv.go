package main

import (
	"math"
	"sort"
	"sync"
)

type KV struct {
	mu    sync.RWMutex // "Read- Write mutex"
	data  map[string][]byte
	vdata map[string][]float64 // for Vector embeddings
}

func NewKV() *KV { // new key-value pair
	return &KV{
		data:  make(map[string][]byte),
		vdata: make(map[string][]float64),
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

// VSet to add a vector
func (kv *KV) VSet(key string, vector []float64) {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	kv.vdata[key] = vector
}

// to hold search results
type VectorMatch struct {
	Key   string
	Score float64
}

// finds the most similar vectors; Uses cosine similarity
func (kv *KV) VSearch(query []float64, limit int) []VectorMatch {
	kv.mu.RLock()
	defer kv.mu.RUnlock()

	var matches []VectorMatch

	for key, vec := range kv.vdata {
		score := cosineSimilarity(query, vec)
		matches = append(matches, VectorMatch{Key: key, Score: score})

	}

	// sort in descending order first
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Score > matches[j].Score
	})

	// Return top 'N' limits
	if limit > len(matches) {
		limit = len(matches)
	}
	return matches[:limit]
}

//Core Math algo 
// Measures the angle between two multi-dimensional vectors
func cosineSimilarity(a, b []float64) float64{
	var dotProduct, normA, normB float64
	for i := 0; i< len(a) && i< len(b); i++ {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]

	}

	if normA == 0 || normB == 0 {
		return 0.0
	}
	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}
