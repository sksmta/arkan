package storage_manager

import (
	js "github.com/sksmta/arkan/formatting"
	"encoding/json"
)


func (sm *StorageManager) WriteJSONBlock(data js.JSONDocument, encrypt bool) (int64, error) {
	sm.writeMutex.Lock()
	defer sm.writeMutex.Unlock()

	blockID := sm.blockIDManager.GetNextID()

	var jsonData []byte
	var err error

	if jsonData, err = js.SerializeJSON(data); err != nil {
		return 0, err
	}

	if encrypt {
		// Encrypt sensitive data before writing
		encryptedData, err := sm.EncryptData(jsonData)
		if err != nil {
			return 0, err
		}
		jsonData = encryptedData
	}

	recordData := serializeRecord(blockID, jsonData)

	// Write to WAL
	if err := sm.wal.WriteRecord(recordData); err != nil {
		return 0, err
	}

	// Write to in-memory cache (store encrypted or unencrypted data)
	sm.cache.Add(blockID, jsonData)

	// Write to file (store encrypted or unencrypted data)
	_, err = sm.file.WriteAt(recordData, blockID*int64(sm.bufferSize))
	if err != nil {
		return 0, err
	}

	return blockID, nil
}

func (sm *StorageManager) ReadJSONBlock(blockID int64, decrypt bool) (js.JSONDocument, error) {
	// Check in-memory cache
	if data, ok := sm.cache.Get(blockID); ok {
		if decrypt {
			// Decrypt data before returning
			decryptedData, err := sm.DecryptData(data)
			if err != nil {
				return nil, err
			}
			return js.DeserializeJSON(decryptedData)
		}
		return js.DeserializeJSON(data)
	}

	data := make([]byte, sm.bufferSize)
	_, err := sm.file.ReadAt(data, blockID*int64(sm.bufferSize))
	if err != nil {
		return nil, err
	}

	// Decrypt data before adding to cache
	if decrypt {
		decryptedData, err := sm.DecryptData(data)
		if err != nil {
			return nil, err
		}
		sm.cache.Add(blockID, decryptedData)
		return js.DeserializeJSON(decryptedData)
	}
	sm.cache.Add(blockID, data)
	return js.DeserializeJSON(data)
}

func (sm *StorageManager) UpdateJSONBlock(blockID int64, newData js.JSONDocument) error {
	sm.writeMutex.Lock()
	defer sm.writeMutex.Unlock()

	var jsonData []byte
	var err error

	if jsonData, err = json.Marshal(newData); err != nil {
		return err
	}

	// Check if the block is encrypted in the cache
	cachedData, ok := sm.cache.Get(blockID)
	if ok {
		// If encrypted, decrypt the cached data before updating
		if decryptedData, err := sm.DecryptData(cachedData); err == nil {
			cachedData = decryptedData
		} else {
			return err
		}
	}

	// Update the JSON data in the block
	updatedData, err := mergeJSON(cachedData, jsonData)

	if err != nil {
		return err
	}

	// Encrypt updated data if needed
	encryptedData, err := sm.EncryptData(updatedData)
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

func mergeJSON(existingData, newData []byte) ([]byte, error) {
	var existingMap map[string]interface{}
	var newMap map[string]interface{}

	if err := json.Unmarshal(existingData, &existingMap); err != nil {
		return nil, err
	}

	if err := json.Unmarshal(newData, &newMap); err != nil {
		return nil, err
	}

	// Merge the two maps (overwriting existing keys with new values)
	for key, value := range newMap {
		existingMap[key] = value
	}

	mergedData, err := json.Marshal(existingMap)
	if err != nil {
		return nil, err
	}

	return mergedData, nil
}
