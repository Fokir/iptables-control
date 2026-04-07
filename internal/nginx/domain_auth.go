package nginx

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	cookieName          = "_sc_domain_auth"
	defaultCookieMaxAge = 2592000 // 30 days in seconds
)

func generateCookieSecret() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func signCookie(domain, secret string, expiresAt time.Time) string {
	expStr := strconv.FormatInt(expiresAt.Unix(), 10)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(domain + "." + expStr))
	sig := hex.EncodeToString(mac.Sum(nil))
	return sig + "." + expStr
}

func verifyCookie(domain, secret, cookieValue string) error {
	parts := strings.SplitN(cookieValue, ".", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid cookie format")
	}

	sig, expStr := parts[0], parts[1]

	expUnix, err := strconv.ParseInt(expStr, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid expiry")
	}

	if time.Now().Unix() > expUnix {
		return fmt.Errorf("cookie expired")
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(domain + "." + expStr))
	expected := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(sig), []byte(expected)) {
		return fmt.Errorf("invalid signature")
	}

	return nil
}
