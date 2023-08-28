package storage_manager

import (
	"os"
	"sync"

	c "github.com/sksmta/arkan/cache"
	wa "github.com/sksmta/arkan/wal"
)

type StorageManager struct {
	file           *os.File
	bufferPool     map[int64][]byte
	bufferSize     int
	cache          *c.LRUCache
	writeMutex     sync.Mutex
	readWriteLock  sync.RWMutex
	wal            *wa.WAL
	key            []byte // AES encryption key
	nonceSize      int    // Nonce size for AES GCM mode
	blockIDManager *BlockIDManager
	encryptionKey  []byte
}
