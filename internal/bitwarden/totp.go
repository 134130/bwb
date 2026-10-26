package bitwarden

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base32"
	"encoding/binary"
	"errors"
	"fmt"
	"hash"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type OTP struct {
	Code      string    `json:"code"`
	ExpiresAt time.Time `json:"expiresAt"`
}

func TOTP(value string, now time.Time) (string, error) {
	result, err := GenerateTOTP(value, now)
	return result.Code, err
}

// GenerateTOTP implements RFC 6238 for Base32 secrets and otpauth://totp URIs.
func GenerateTOTP(value string, now time.Time) (OTP, error) {
	if len(value) == 0 || len(value) > 10000 || now.Unix() < 0 {
		return OTP{}, errors.New("invalid TOTP secret or time")
	}
	secret, algorithm, digits, period := value, "SHA1", 6, 30
	if strings.Contains(value, ":") {
		u, err := url.Parse(value)
		if err != nil || u.Scheme != "otpauth" || u.Host != "totp" || u.User != nil || u.Fragment != "" {
			return OTP{}, errors.New("TOTP requires a Base32 secret or otpauth://totp URI")
		}
		query, err := url.ParseQuery(u.RawQuery)
		if err != nil {
			return OTP{}, errors.New("invalid TOTP URI parameters")
		}
		for _, key := range []string{"secret", "algorithm", "digits", "period"} {
			if values, exists := query[key]; exists && (len(values) != 1 || values[0] == "") {
				return OTP{}, errors.New("invalid TOTP URI parameters")
			}
		}
		secret = query.Get("secret")
		if query.Has("algorithm") {
			algorithm = strings.ToUpper(query.Get("algorithm"))
		}
		if query.Has("digits") {
			digits, err = strconv.Atoi(query.Get("digits"))
			if err != nil {
				return OTP{}, errors.New("TOTP digits must be 6 or 8")
			}
		}
		if query.Has("period") {
			period, err = strconv.Atoi(query.Get("period"))
			if err != nil {
				return OTP{}, errors.New("TOTP period must be between 1 and 86400 seconds")
			}
		}
	}
	if digits != 6 && digits != 8 {
		return OTP{}, errors.New("TOTP digits must be 6 or 8")
	}
	if period < 1 || period > 86400 {
		return OTP{}, errors.New("TOTP period must be between 1 and 86400 seconds")
	}
	var digest func() hash.Hash
	switch algorithm {
	case "SHA1":
		digest = sha1.New
	case "SHA256":
		digest = sha256.New
	case "SHA512":
		digest = sha512.New
	default:
		return OTP{}, errors.New("TOTP supports SHA1, SHA256 and SHA512")
	}
	secret = strings.ToUpper(strings.Join(strings.Fields(secret), ""))
	var key []byte
	var err error
	if strings.Contains(secret, "=") {
		key, err = base32.StdEncoding.DecodeString(secret)
	} else {
		key, err = base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(secret)
	}
	defer clear(key)
	if err != nil || len(key) == 0 {
		return OTP{}, errors.New("invalid Base32 TOTP secret")
	}
	var counter [8]byte
	binary.BigEndian.PutUint64(counter[:], uint64(now.Unix()/int64(period)))
	mac := hmac.New(digest, key)
	_, _ = mac.Write(counter[:])
	sum := mac.Sum(nil)
	defer clear(sum)
	offset := sum[len(sum)-1] & 15
	number := binary.BigEndian.Uint32(sum[offset:offset+4]) & 0x7fffffff
	modulus := uint32(1000000)
	if digits == 8 {
		modulus = 100000000
	}
	return OTP{Code: fmt.Sprintf("%0*d", digits, number%modulus), ExpiresAt: time.Unix((now.Unix()/int64(period)+1)*int64(period), 0)}, nil
}
