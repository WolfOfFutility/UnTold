package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	random "math/rand/v2"
	"os"
	"strings"
	"log"
)

// Encrypt the specified data with a specified key
func encrpytData(key []byte, data []byte) ([]byte, error) {
	aes, err := aes.NewCipher(key)
    if err != nil {
        return nil, err
    }

    gcm, err := cipher.NewGCM(aes)
    if err != nil {
        return nil, err
    }

    // We need a 12-byte nonce for GCM (modifiable if you use cipher.NewGCMWithNonceSize())
    // A nonce should always be randomly generated for every encryption.
    nonce := make([]byte, gcm.NonceSize())
    _, err = rand.Read(nonce)
    if err != nil {
        return nil, err
    }

    // ciphertext here is actually nonce+ciphertext
    // So that when we decrypt, just knowing the nonce size
    // is enough to separate it from the ciphertext.
    ciphertext := gcm.Seal(nonce, nonce, []byte(data), nil)

    return ciphertext, nil
}

// Decrypt the specified data with a specified key
func decryptData(key []byte, data []byte) ([]byte, error) {
	aes, err := aes.NewCipher(key)
    if err != nil {
        return nil, err
    }

    gcm, err := cipher.NewGCM(aes)
    if err != nil {
        return nil, err
    }

    // Since we know the ciphertext is actually nonce+ciphertext
    // And len(nonce) == NonceSize(). We can separate the two.
    nonceSize := gcm.NonceSize()
    nonce, data := data[:nonceSize], data[nonceSize:]

    plaintext, err := gcm.Open(nil, []byte(nonce), []byte(data), nil)
    if err != nil {
        return nil, err
    }

    return plaintext, nil
}

// Either get an existing key or generate a new one
func generateEncryptionKey() (error) {
	content, err := os.ReadFile("keys/main.dat")
	if err != nil {
		if strings.Contains(err.Error(), "The system cannot find the file") {
			log.Println("System cannot find an existing key - creating a new one.")
		} else {
			return err
		}
	}

	if len(content) < 32 {
		var letters = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

		// Handle opening / creating the new key file
		file, err := os.OpenFile("keys/main.dat", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
		if err != nil {
			return err
		}

		// Generate a new byte key at random
		b := make([]byte, 32)
		for i := range b {
			b[i] = letters[random.IntN(len(letters))]
		}
		newKey := string(b)

		// Set the key to environment
		setEnvErr := os.Setenv("EK", newKey)
		if setEnvErr != nil {
			return setEnvErr
		}

		// Handle errors with writing the key to file
		_, fileWriteErr := file.Write(b)
		if fileWriteErr != nil {
			return fileWriteErr
		}

		defer file.Close()
	} else {
		// Set the key to environment
		setEnvErr := os.Setenv("EK", string(content))
		if setEnvErr != nil {
			return setEnvErr
		}
	}

	return nil
}

