package bitwarden

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hkdf"
	"crypto/hmac"
	"crypto/pbkdf2"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"strings"

	"golang.org/x/crypto/argon2"
)

var (
	ErrUnsupported = errors.New("unsupported vault format or authentication method")
	ErrCipher      = errors.New("invalid or unauthenticated ciphertext")
)

type kdfConfig struct {
	Type, Iterations, Memory, Parallelism int
}

func (k kdfConfig) validate() error {
	switch k.Type {
	case 0:
		if k.Iterations < 100000 || k.Iterations > 2000000 {
			return errors.New("PBKDF2 iterations must be between 100000 and 2000000")
		}
	case 1:
		if k.Iterations < 2 || k.Iterations > 10 || k.Memory < 16 || k.Memory > 1024 || k.Parallelism < 1 || k.Parallelism > 16 {
			return errors.New("Argon2id requires 2–10 iterations, 16–1024 MiB memory and 1–16 parallelism")
		}
	default:
		return ErrUnsupported
	}
	return nil
}

// Serialize memory-heavy derivations across profiles; waiting remains cancellable.
var argonWork = make(chan struct{}, 1)

func derive(ctx context.Context, email string, password []byte, kdf kdfConfig) ([]byte, string, error) {
	if err := kdf.validate(); err != nil {
		return nil, "", err
	}
	if err := ctx.Err(); err != nil {
		return nil, "", err
	}
	var master []byte
	var err error
	if kdf.Type == 1 {
		select {
		case argonWork <- struct{}{}:
		case <-ctx.Done():
			return nil, "", ctx.Err()
		}
		defer func() { <-argonWork }()
		if err := ctx.Err(); err != nil {
			return nil, "", err
		}
		salt := sha256.Sum256([]byte(email))
		// IDKey cannot be interrupted; observe cancellation immediately after computation.
		master = argon2.IDKey(password, salt[:], uint32(kdf.Iterations), uint32(kdf.Memory)*1024, uint8(kdf.Parallelism), 32)
	} else {
		master, err = pbkdf2.Key(sha256.New, string(password), []byte(email), kdf.Iterations, 32)
		if err != nil {
			return nil, "", err
		}
	}
	defer clear(master)
	if err := ctx.Err(); err != nil {
		return nil, "", err
	}
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
