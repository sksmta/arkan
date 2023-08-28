package wal

import (
	"errors"
	"io"
	"os"
	"sync"
)

const maxRecordSize = 8192 // Adjust as needed

type WAL struct {
	file  *os.File
	mutex sync.Mutex
}

func NewWAL(filePath string) (*WAL, error) {
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_RDWR, os.ModePerm)

	if err != nil {
		return nil, err
	}

	return &WAL{
		file: file,
	}, nil
}

func (w *WAL) WriteRecord(recordData []byte) error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	_, err := w.file.Write(recordData)
	if err != nil {
		return err
	}
	return nil
}

func (w *WAL) ReadRecords() ([][]byte, error) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	_, err := w.file.Seek(0, 0)
	if err != nil {
		return nil, err
	}

	var records [][]byte
	record := make([]byte, maxRecordSize)

	for {
		bytesRead, err := w.file.Read(record)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
		records = append(records, record[:bytesRead])
	}

	return records, nil
}

func (w *WAL) Close() {
	err := w.file.Close()
	if err != nil {
		return
	}
}
