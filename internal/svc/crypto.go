package svc

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"io"
)

func hash(key string) string {
	hash := sha256.Sum256([]byte(key))
	hex := hex.EncodeToString(hash[:])
	return hex
}

func encrypt(key, msg string) (string, error) {
	bkey := sha256.Sum256([]byte(key))
	bmsg := []byte(msg)
	c, err := aes.NewCipher(bkey[:])
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	_, err = io.ReadFull(rand.Reader, nonce)
	if err != nil {
		return "", err
	}
	encrypted := gcm.Seal(nonce, nonce, bmsg, nil)
	return hex.EncodeToString(encrypted), nil
}

func decrypt(key, msg string) (string, error) {
	bkey := sha256.Sum256([]byte(key))
	bencr, err := hex.DecodeString(msg)
	if err != nil {
		return "", err
	}
	c, err := aes.NewCipher(bkey[:])
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return "", err
	}
	nonceSize := gcm.NonceSize()
	nonce, encr := bencr[:nonceSize], bencr[nonceSize:]
	decr, err := gcm.Open(nil, nonce, encr, nil)
	if err != nil {
		return "", err
	}
	return string(decr), nil
}
