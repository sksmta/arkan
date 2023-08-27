package storage_manager

import (
	_ "fmt"
	"os"
	"sync"

	c "github.com/sksmta/arkan/cache"
	wa "github.com/sksmta/arkan/wal"
)

const (
	blockSize    = 4096
	maxCacheSize = 100
)

type Record struct {
	BlockID int64
	Data    []byte
}

type StorageManager struct {
	file          *os.File
	bufferPool    map[int64][]byte
	bufferSize    int
	cache         *c.LRUCache
	writeMutex    sync.Mutex
	readWriteLock sync.RWMutex
	wal           *wa.WAL
}

func createStorageManager(filePath string) (*StorageManager, error) {
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}

	cache := c.NewLRUCache(maxCacheSize)

	return &StorageManager{
		file:       file,
		bufferPool: make(map[int64][]byte),
		bufferSize: blockSize,
		cache:      cache,
	}, nil
}

func NewStorageManager(filePath, walFilePath string) (*StorageManager, error) {
	// Create the storage manager and cache
	sm, err := createStorageManager(filePath)
	if err != nil {
		return nil, err
	}

	// Open the WAL and read records for recovery
	wal, err := wa.NewWAL(walFilePath)
	if err != nil {
		return nil, err
	}
	sm.wal = wal // Assign WAL to the storage manager

	// Recover records from WAL if they don't already exist in the database
	if err := sm.recoverFromWAL(); err != nil {
		return nil, err
	}

	return sm, nil
}

func (sm *StorageManager) ReadBlock(blockID int64) ([]byte, error) {
	sm.readWriteLock.RLock()
	defer sm.readWriteLock.RUnlock()

	cachedData, ok := sm.cache.Get(blockID)
	if ok {
		return cachedData, nil
	}

	data := make([]byte, sm.bufferSize)
	_, err := sm.file.ReadAt(data, blockID*int64(sm.bufferSize))
	if err != nil {
		return nil, err
	}

	sm.cache.Add(blockID, data)
	return data, nil
}

func (sm *StorageManager) WriteBlock(blockID int64, data []byte) error {
	sm.writeMutex.Lock()
	defer sm.writeMutex.Unlock()

	recordData := serializeRecord(blockID, data)

	// Write to WAL
	if err := sm.wal.WriteRecord(recordData); err != nil {
		return err
	}

	// Write to in-memory cache
	sm.cache.Add(blockID, data)

	// Write to file
	_, err := sm.file.WriteAt(recordData, blockID*int64(sm.bufferSize))
	if err != nil {
		return err
	}

	return nil
}

func (sm *StorageManager) close() {
	err := sm.file.Close()
	if err != nil {
		return
	}
	sm.wal.Close()
}

func serializeRecord(blockID int64, data []byte) []byte {
	// Use variable-length integer encoding for the block ID
	blockIDBytes := encodeVarInt64(blockID)

	recordSize := len(blockIDBytes) + len(data)
	record := make([]byte, recordSize)
	copy(record, blockIDBytes)
	copy(record[len(blockIDBytes):], data)
	return record
}

func encodeVarInt64(value int64) []byte {
	var encoded []byte

	for value >= 0x80 {
		encoded = append(encoded, byte(value)|0x80)
		value >>= 7
	}

	encoded = append(encoded, byte(value))
	return encoded
}

func (sm *StorageManager) recoverFromWAL() error {
	walRecords, err := sm.wal.ReadRecords()
	if err != nil {
		return err
	}

	for _, recordData := range walRecords {
		blockID, data := wa.DeserializeRecord(recordData)
		if err := sm.WriteBlock(blockID, data); err != nil {
			return err
		}
	}

	return nil
}
