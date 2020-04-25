package utils

import (
	"crypto/aes"
	"crypto/cipher"
	crand "crypto/rand"
	"errors"
	"io"
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func EncryptAES(key []byte, s []byte) (encmess []byte, err error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return
	}

	c := make([]byte, aes.BlockSize+len(s))
	iv := c[:aes.BlockSize]
	if _, err = io.ReadFull(crand.Reader, iv); err != nil {
		return
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(c[aes.BlockSize:], s)

	return c, nil
}

func DecryptAES(key []byte, s []byte) (decodedmess []byte, err error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return
	}

	if len(s) < aes.BlockSize {
		err = errors.New("Ciphertext block size is too short!")
		return
	}

	iv := s[:aes.BlockSize]
	s = s[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(s, s)

	return s, nil
}
