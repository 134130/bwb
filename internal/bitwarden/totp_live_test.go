package bitwarden

import (
	"encoding/base32"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestLiveTOTPRFC6238(t *testing.T) {
	cases := []struct {
		unix  int64
		codes [3]string
	}{
		{59, [3]string{"94287082", "46119246", "90693936"}},
		{1111111109, [3]string{"07081804", "68084774", "25091201"}},
		{1111111111, [3]string{"14050471", "67062674", "99943326"}},
		{1234567890, [3]string{"89005924", "91819424", "93441116"}},
		{2000000000, [3]string{"69279037", "90698825", "38618901"}},
		{20000000000, [3]string{"65353130", "77737706", "47863826"}},
	}
	for _, c := range cases {
		for i, algorithm := range []string{"SHA1", "SHA256", "SHA512"} {
			secret := strings.Repeat("1234567890", 7)[:[]int{20, 32, 64}[i]]
			uri := fmt.Sprintf("otpauth://totp/test?secret=%s&algorithm=%s&digits=8", base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString([]byte(secret)), algorithm)
			result, err := GenerateTOTP(uri, time.Unix(c.unix, 0))
			if err != nil || result.Code != c.codes[i] || result.ExpiresAt.Unix() != (c.unix/30+1)*30 {
				t.Fatalf("%s at %d: %v %v", algorithm, c.unix, result, err)
			}
		}
	}
}

func TestLiveTOTPBoundaryAndInvalidSettings(t *testing.T) {
	seed := "GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ"
	a, err := GenerateTOTP(seed, time.Unix(59, 999999999))
	b, nextErr := GenerateTOTP(seed, time.Unix(60, 0))
	if err != nil || nextErr != nil || a.Code != "287082" || a.Code == b.Code || a.ExpiresAt.Unix() != 60 {
		t.Fatal(a, b, err, nextErr)
	}
	for _, value := range []string{"", "secret!", "steam://test", "otpauth://hotp/x?secret=" + seed,
		"otpauth://totp/x?secret=" + seed + "&period=0", "otpauth://totp/x?secret=" + seed + "&digits=9",
		"otpauth://totp/x?secret=" + seed + "&algorithm=MD5", "otpauth://totp/x?secret=" + seed + "&period=30&period=60"} {
		_, err := GenerateTOTP(value, time.Unix(59, 0))
		if err == nil || strings.Contains(err.Error(), seed) {
			t.Fatal("accepted invalid settings or exposed seed", err)
		}
	}
	result, err := GenerateTOTP("otpauth://totp/x?secret="+seed+"&period=60", time.Unix(59, 0))
	if err != nil || result.ExpiresAt.Unix() != 60 {
		t.Fatal(result, err)
	}
}
