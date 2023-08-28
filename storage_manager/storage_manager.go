package storage_manager

import (
	"crypto/rand"
	_ "fmt"
	"os"

	c "github.com/sksmta/arkan/cache"
	wa "github.com/sksmta/arkan/wal"
)

const (
	blockSize    = 4096
	maxCacheSize = 100

	keySize = 32
)

type Record struct {
	BlockID int64
	Data    []byte
}

func createStorageManager(filePath string) (*StorageManager, error) {
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_RDWR, os.ModePerm)
	if err != nil {
		return nil, err
	}

	cache := c.NewLRUCache(maxCacheSize)

	// Generate a new keyString
	key := make([]byte, keySize)
	_, err = rand.Read(key)
	if err != nil {
		return nil, err
	}

	return &StorageManager{
		file:          file,
		bufferPool:    make(map[int64][]byte),
		bufferSize:    blockSize,
		cache:         cache,
	}, nil
}

func NewStorageManager(filePath, walFilePath string) (*StorageManager, error) {
	// Create the storage manager and cache
	sm, err := createStorageManager(filePath)
	if err != nil {
		return nil, err
	}

	// Initialize block ID manager with the last used block ID + 1
	lastUsedBlockID, err := sm.getLastUsedBlockID()
	if err != nil {
		return nil, err
	}
	sm.blockIDManager = NewBlockIDManager(lastUsedBlockID + 1)

	// Generate encryption key
	if err := sm.generateEncryptionKey(); err != nil {
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
		// Decrypt cached data before returning
		decryptedData, err := sm.DecryptData(cachedData)
		if err != nil {
			return nil, err
		}
		return decryptedData, nil
	}

	data := make([]byte, sm.bufferSize)
	_, err := sm.file.ReadAt(data, blockID*int64(sm.bufferSize))
	if err != nil {
		return nil, err
	}

	// Decrypt data before adding to cache
	decryptedData, err := sm.DecryptData(data)
	if err != nil {
		return nil, err
	}

	sm.cache.Add(blockID, decryptedData)
	return decryptedData, nil
}

func (sm *StorageManager) WriteBlock(data []byte) error {
	sm.writeMutex.Lock()
	defer sm.writeMutex.Unlock()

	blockID := sm.blockIDManager.GetNextID()

	// Encrypt data before writing
	encryptedData, err := sm.EncryptData(data)
	if err != nil {
		return err
	}

	recordData := serializeRecord(blockID, encryptedData)

	// Write to WAL
	if err := sm.wal.WriteRecord(recordData); err != nil {
		return err
	}

	// Write to in-memory cache (store encrypted data)
	sm.cache.Add(blockID, encryptedData)

	// Write to file (store encrypted data)
	_, err = sm.file.WriteAt(recordData, blockID*int64(sm.bufferSize))
	if err != nil {
		return err
	}

	return nil
}

func (sm *StorageManager) Close() {
	err := sm.file.Close()
	if err != nil {
		return
	}
	println("Closing WAL")
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
		_, data := wa.DeserializeRecord(recordData)
		if err := sm.WriteBlock(data); err != nil {
			return err
		}
	}

	return nil
}

func (sm *StorageManager) generateEncryptionKey() error {
	key := make([]byte, 32) // You can adjust the key size as needed
	_, err := rand.Read(key)
	if err != nil {
		return err
	}
	sm.encryptionKey = key
	return nil
}
