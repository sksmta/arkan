package storage_manager

import (
	"encoding/json"
	"os"
	"testing"
)

type TestData struct {
	ID   int64
	Name string
}

func TestStorageManager_WriteAndReadBlock(t *testing.T) {
	// Create a temporary test file

	tempDir := os.TempDir()

	tmpFile, err := os.CreateTemp(tempDir, "testdb")
	if err != nil {
		t.Fatal(err)
	}
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {

		}
	}(tmpFile.Name())

	// Create a new StorageManager for testing
	sm, err := NewStorageManager(tmpFile.Name(), "test_wal.log")
	if err != nil {
		t.Fatal(err)
	}
	defer sm.close()

	// Create test data
	data := TestData{
		ID:   123,
		Name: "John Doe",
	}

	// Convert test data to bytes
	testDataBytes, err := serializeTestRecord(data)
	if err != nil {
		t.Fatal(err)
	}

	blockID := int64(0)
	err = sm.WriteBlock(blockID, testDataBytes)
	if err != nil {
		t.Fatal(err)
	}

	// Read the written block
	readDataBytes, err := sm.ReadBlock(blockID)
	if err != nil {
		t.Fatal(err)
	}

	// Deserialize the read data
	readData, err := deserializeTestRecord(readDataBytes)
	if err != nil {
		t.Fatal(err)
	}

	// Compare read data with original data
	if readData.ID != data.ID || readData.Name != data.Name {
		t.Errorf("Read data does not match expected data")
	}
}

func serializeTestRecord(data TestData) ([]byte, error) {
	serializedData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return serializedData, nil
}

func deserializeTestRecord(data []byte) (TestData, error) {
	var testData TestData
	err := json.Unmarshal(data, &testData)
	if err != nil {
		return TestData{}, err
	}
	return testData, nil
}
