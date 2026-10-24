package bitwarden

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hkdf"
	"crypto/hmac"
	"crypto/pbkdf2"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"strings"
)

var (
	ErrUnsupported = errors.New("unsupported vault format or authentication method")
	ErrCipher      = errors.New("invalid or unauthenticated ciphertext")
)

// derive supports PBKDF2 accounts only. The bound limits work requested by a server.
func derive(email string, password []byte, iterations int) ([]byte, string, error) {
	if iterations < 100000 || iterations > 2000000 {
		return nil, "", errors.New("PBKDF2 iterations must be between 100000 and 2000000")
	}
	master, err := pbkdf2.Key(sha256.New, string(password), []byte(email), iterations, 32)
	if err != nil {
		return nil, "", err
	}
	defer clear(master)
	enc, err := hkdf.Expand(sha256.New, master, "enc", 32)
	if err != nil {
		return nil, "", err
	}
	defer clear(enc)
	mac, err := hkdf.Expand(sha256.New, master, "mac", 32)
	if err != nil {
		return nil, "", err
	}
	defer clear(mac)
	hash, err := pbkdf2.Key(sha256.New, string(master), password, 1, 32)
	if err != nil {
		return nil, "", err
	}
	defer clear(hash)
	return append(enc[:len(enc):len(enc)], mac...), base64.StdEncoding.EncodeToString(hash), nil
}

func decrypt(value string, key []byte) ([]byte, error) {
	if len(key) != 64 {
		return nil, ErrCipher
	}
	if !strings.HasPrefix(value, "2.") {
		return nil, ErrUnsupported
	}
	parts := strings.Split(value[2:], "|")
	if len(parts) != 3 {
		return nil, ErrCipher
	}
	decoded := make([][]byte, 3)
	for i, part := range parts {
		var err error
		decoded[i], err = base64.StdEncoding.Strict().DecodeString(part)
		if err != nil {
			return nil, ErrCipher
		}
	}
	iv, ciphertext, mac := decoded[0], decoded[1], decoded[2]
	if len(iv) != aes.BlockSize || len(ciphertext) == 0 || len(ciphertext)%aes.BlockSize != 0 || len(mac) != sha256.Size {
		return nil, ErrCipher
	}
	auth := hmac.New(sha256.New, key[32:])
	auth.Write(iv)
	auth.Write(ciphertext)
	if !hmac.Equal(mac, auth.Sum(nil)) {
		return nil, ErrCipher
	}
	block, err := aes.NewCipher(key[:32])
	if err != nil {
		return nil, ErrCipher
	}
	plain := make([]byte, len(ciphertext))
	cipher.NewCBCDecrypter(block, iv).CryptBlocks(plain, ciphertext)
	n := int(plain[len(plain)-1])
	if n < 1 || n > aes.BlockSize {
		clear(plain)
		return nil, ErrCipher
	}
	for _, b := range plain[len(plain)-n:] {
		if int(b) != n {
			clear(plain)
			return nil, ErrCipher
		}
	}
	clear(plain[len(plain)-n:])
	return plain[:len(plain)-n], nil
}
