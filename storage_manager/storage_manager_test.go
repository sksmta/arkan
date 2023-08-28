package storage_manager

import (
	"testing"
)

func TestStorageManager_WriteAndReadBlock(t *testing.T) {
	// Create an instance of your StorageManager for testing
	sm, err := NewStorageManager("test_data.db", "wal_test.log")
	if err != nil {
		t.Fatalf("Failed to create StorageManager: %v", err)
	}

	// Write a block
	data := []byte("Hello, World!")
	err = sm.WriteBlock(0, data)
	if err != nil {
		t.Errorf("Error writing block: %v", err)
	}

	// Read the block
	readData, err := sm.ReadBlock(0)
	if err != nil {
		t.Errorf("Error reading block: %v", err)
	}

	// Compare read data with original data
	if string(readData) != string(data) {
		t.Errorf("Read data doesn't match original data")
	} else {
		println(string(readData))
		println("Read data matches original data")
	}
}
