package bitwarden

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/awnumar/memguard"
)

// Golden values were generated independently by the Python snapshot builder.
func TestDeriveGolden(t *testing.T) {
	key, hash, err := derive("bwb-integration@example.invalid", []byte("PUBLIC-INTEGRATION-ONLY-not-a-real-password!"), 600000)
	if err != nil {
		t.Fatal(err)
	}
	defer clear(key)
	if hex.EncodeToString(key) != "0da2649fd6b2e781d5ae194fdf07693e03a644792112102759b85eae9994e385c1d0595858f757265d229ff771c04900cc37f8d5ad0a9b90c6a75eba2296b4d7" || hash != "Jh9Q9wvoW60Whc2Yytb/Pw3kFfgikGUG1dj/IK9nnxw=" {
		t.Fatal("KDF mismatch")
	}
	for _, iterations := range []int{-1, 0, 99999, 2000001} {
		if _, _, err := derive("user@example.invalid", []byte("password"), iterations); err == nil {
			t.Fatalf("accepted iterations %d", iterations)
		}
	}
}

func TestDecryptGoldenAndMalformed(t *testing.T) {
	key := make([]byte, 64)
	for i := range key {
		key[i] = byte(i)
	}
	const golden = "2.a5KHrY2pt1G+DeKpMvHM9A==|aCnNh9cQDXrVQcuZ6M00ww==|k9HNeXW3D5IV3iqsDB/5o15fzkdmUs4vx1a4XkDypYk="
	plain, err := decrypt(golden, key)
	if err != nil || string(plain) != "한글\nsecret" {
		t.Fatal("golden ciphertext failed", err)
	}
	clear(plain)
	for _, value := range []string{"", "0.x", "2.", "2.a|b|c", golden + "|extra", strings.Replace(golden, "aCnN", "bCnN", 1), strings.Replace(golden, "a5KH", "b5KH", 1)} {
		plain, err := decrypt(value, key)
		if err == nil || plain != nil {
			t.Fatal("accepted malformed or tampered ciphertext")
		}
	}
	for _, key := range [][]byte{nil, make([]byte, 32), make([]byte, 64)} {
		plain, err := decrypt(golden, key)
		if !errors.Is(err, ErrCipher) || plain != nil {
			t.Fatal("accepted wrong key")
		}
	}
	// A correct MAC is insufficient: invalid PKCS#7 must still be rejected.
	for _, padded := range [][]byte{make([]byte, 16), bytes.Repeat([]byte{17}, 16), append(bytes.Repeat([]byte{0}, 15), 2)} {
		plain, err := decrypt(encryptPadded(t, padded, key), key)
		if err == nil || plain != nil {
			t.Fatal("accepted invalid padding")
		}
	}
	empty, err := decrypt(encryptPadded(t, bytes.Repeat([]byte{16}, 16), key), key)
	if err != nil || len(empty) != 0 {
		t.Fatal("empty plaintext failed")
	}
}

func encryptPadded(t *testing.T, padded, key []byte) string {
	t.Helper()
	iv := make([]byte, 16)
	block, err := aes.NewCipher(key[:32])
	if err != nil {
		t.Fatal(err)
	}
	out := make([]byte, len(padded))
	cipher.NewCBCEncrypter(block, iv).CryptBlocks(out, padded)
	mac := hmac.New(sha256.New, key[32:])
	mac.Write(iv)
	mac.Write(out)
	return "2." + base64.StdEncoding.EncodeToString(iv) + "|" + base64.StdEncoding.EncodeToString(out) + "|" + base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func TestUnsupportedItemPolicies(t *testing.T) {
	key := bytes.Repeat([]byte{1}, 64)
	session := &Session{key: memguard.NewBufferFromBytes(bytes.Clone(key))}
	defer session.Close()
	name := "PRIVATE-NAME"
	padding := 16 - len(name)%16
	encrypted := encryptPadded(t, append([]byte(name), bytes.Repeat([]byte{byte(padding)}, padding)...), key)
	for _, tc := range []struct {
		change func(map[string]any)
		want   string
	}{
		{func(m map[string]any) { m["type"] = 9 }, "item type 9"},
		{func(m map[string]any) { m["reprompt"] = 2 }, "reprompt type 2"},
		{func(m map[string]any) { m["fields"] = []any{map[string]any{"type": 4}} }, "custom field type 4"},
		{func(m map[string]any) {
			m["login"] = map[string]any{"fido2Credentials": []any{map[string]any{"keyValue": "PRIVATE-KEY"}}}
		}, "passkey credentials"},
	} {
		t.Run(tc.want, func(t *testing.T) {
			item := map[string]any{"id": "20000000-0000-4000-8000-000000000001", "type": 1, "name": encrypted, "login": map[string]any{}}
			tc.change(item)
			raw, _ := json.Marshal(item)
			_, err := session.decryptSnapshot(t.Context(), syncResponse{Ciphers: []json.RawMessage{raw}})
			if !errors.Is(err, ErrUnsupported) || !strings.Contains(err.Error(), "decrypt vault item 1: "+tc.want) {
				t.Fatal("missing item diagnostic", err)
			}
			if strings.Contains(err.Error(), "PRIVATE") || strings.Contains(err.Error(), encrypted) {
				t.Fatal("item diagnostic leaked vault contents")
			}
		})
	}
}

func TestReadMissingEmptyAndAmbiguous(t *testing.T) {
	name, empty := "token", ""
	id := "20000000-0000-4000-8000-000000000001"
	vault := Vault{Items: []Item{{ID: id, Type: 1, Data: map[string]any{"password": "", "username": nil, "future": "2.encrypted"}, Fields: []Field{{Name: &name, Value: &empty}}}}}
	for _, path := range []string{"login/password", "fields/token"} {
		value, err := vault.Read(id, path)
		if err != nil || value != "" {
			t.Fatal("empty value not preserved")
		}
	}
	for _, path := range []string{"login/username", "login/missing", "login/future", "card/number", "notes"} {
		if _, err := vault.Read(id, path); !errors.Is(err, ErrNotFound) {
			t.Fatal("missing field was not rejected", path)
		}
	}
	vault.Items[0].Fields = append(vault.Items[0].Fields, vault.Items[0].Fields[0])
	if _, err := vault.Read(id, "fields/token"); err == nil {
		t.Fatal("duplicate field selected silently")
	}
}

func TestReadCustomFieldWithSlashes(t *testing.T) {
	id := "20000000-0000-4000-8000-000000000001"
	name, value := "service/token", "secret"
	vault := Vault{Items: []Item{{ID: id, Type: 1, Fields: []Field{{Name: &name, Value: &value}}}}}
	got, err := vault.Read(id, "fields/service/token")
	if err != nil || got != value {
		t.Fatal("slash-containing field not found", err)
	}
	if _, err := vault.Read(id, "fields/service"); !errors.Is(err, ErrNotFound) {
		t.Fatal("partial name matched", err)
	}
	vault.Items[0].Fields = append(vault.Items[0].Fields, vault.Items[0].Fields[0])
	if _, err := vault.Read(id, "fields/service/token"); err == nil {
		t.Fatal("duplicate slash-containing field selected")
	}
}
