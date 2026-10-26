package bitwarden

import (
	"encoding/base32"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestTOTPRFC6238(t *testing.T) {
	secrets := []string{"12345678901234567890", "12345678901234567890123456789012", "1234567890123456789012345678901234567890123456789012345678901234"}
	algorithms := []string{"SHA1", "SHA256", "SHA512"}
	for _, tc := range []struct {
		unix  int64
		codes [3]string
	}{
		{59, [3]string{"94287082", "46119246", "90693936"}},
		{1111111109, [3]string{"07081804", "68084774", "25091201"}},
		{1111111111, [3]string{"14050471", "67062674", "99943326"}},
		{1234567890, [3]string{"89005924", "91819424", "93441116"}},
		{2000000000, [3]string{"69279037", "90698825", "38618901"}},
		{20000000000, [3]string{"65353130", "77737706", "47863826"}},
	} {
		for i, algorithm := range algorithms {
			seed := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString([]byte(secrets[i]))
			uri := fmt.Sprintf("otpauth://totp/test?secret=%s&algorithm=%s&digits=8", seed, algorithm)
			code, err := TOTP(uri, time.Unix(tc.unix, 0))
			if err != nil || code != tc.codes[i] {
				t.Fatalf("%d/%s: %q, %v", tc.unix, algorithm, code, err)
			}
		}
	}
	seed := base32.StdEncoding.EncodeToString([]byte(secrets[0]))
	if code, err := TOTP(strings.ToLower(seed[:8]+" "+seed[8:]), time.Unix(59, 0)); err != nil || code != "287082" {
		t.Fatal(code, err)
	}
	if code, err := TOTP("otpauth://totp/test?secret="+seed+"&period=60", time.Unix(59, 0)); err != nil || code != "755224" {
		t.Fatal(code, err)
	}
}

func TestTOTPRejectsMalformedSecretsWithoutEchoingThem(t *testing.T) {
	seed := "GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ"
	for _, value := range []string{"", "PRIVATE-INVALID-SECRET", "steam://PRIVATE", "otpauth://hotp/test?secret=" + seed,
		"otpauth://totp/test?secret=" + seed + "&secret=OTHER", "otpauth://totp/test?secret=", "otpauth://totp/test?algorithm=SHA1",
		"otpauth://totp/test?secret=" + seed + "&digits=7", "otpauth://totp/test?secret=" + seed + "&period=0",
		"otpauth://totp/test?secret=" + seed + "&period=86401", "otpauth://totp/test?secret=" + seed + "&algorithm=MD5",
		"otpauth://totp/test?secret=%zz", "otpauth://user@totp/test?secret=" + seed, "otpauth://totp/test?secret=" + seed + "&digits=",
	} {
		code, err := TOTP(value, time.Unix(59, 0))
		if err == nil || code != "" || strings.Contains(err.Error(), "PRIVATE") || strings.Contains(err.Error(), seed) {
			t.Fatal("invalid secret accepted or echoed", err)
		}
	}
	if _, err := TOTP(seed, time.Unix(-1, 0)); err == nil {
		t.Fatal("negative time accepted")
	}
}
