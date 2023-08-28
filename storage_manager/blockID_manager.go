package storage_manager

import (
	"sync"
)

type BlockIDManager struct {
	currentID int64
	mutex     sync.Mutex
}

func NewBlockIDManager(initialID int64) *BlockIDManager {
	return &BlockIDManager{
		currentID: initialID,
	}
}

func (bm *BlockIDManager) GetNextID() int64 {
	bm.mutex.Lock()
	defer bm.mutex.Unlock()
	id := bm.currentID
	bm.currentID++
	return id
}

func (sm *StorageManager) getLastUsedBlockID() (int64, error) {
	// Seek to the end of the file to find the last written record
	fileInfo, err := sm.file.Stat()
	if err != nil {
		return 0, err
	}

	fileSize := fileInfo.Size()
	if fileSize == 0 {
		// No records written yet
		return 0, nil
	}

	// Calculate the block ID of the last written record
	lastRecordBlockID := (fileSize - 1) / int64(sm.bufferSize)

	return lastRecordBlockID, nil
}