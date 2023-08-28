package storage_manager

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
)

func (sm *StorageManager) EncryptData(data []byte) ([]byte, error) {
	nonce := make([]byte, sm.nonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(sm.key)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	encryptedData := aesgcm.Seal(nil, nonce, data, nil)
	return append(nonce, encryptedData...), nil
}

// DecryptData decrypts the given AES-GCM encrypted data
func (sm *StorageManager) DecryptData(encryptedData []byte) ([]byte, error) {
	nonce := encryptedData[:sm.nonceSize]
	ciphertext := encryptedData[sm.nonceSize:]

	block, err := aes.NewCipher(sm.key)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	decryptedData, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return decryptedData, nil
}