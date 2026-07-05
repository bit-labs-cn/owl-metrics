package signing

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"strings"
	"time"
)

const maxSkew = 5 * time.Minute

// VerifyHMAC 校验 HMAC-SHA256 签名。
// payload = timestamp + "\n" + body
func VerifyHMAC(secret, timestamp, signature, body string) bool {
	secret = strings.TrimSpace(secret)
	timestamp = strings.TrimSpace(timestamp)
	signature = strings.TrimSpace(signature)
	if secret == "" || timestamp == "" || signature == "" {
		return false
	}
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return false
	}
	now := time.Now().UnixMilli()
	if ts < now-int64(maxSkew.Milliseconds()) || ts > now+int64(maxSkew.Milliseconds()) {
		return false
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(timestamp))
	mac.Write([]byte("\n"))
	mac.Write([]byte(body))
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(strings.ToLower(signature)), []byte(strings.ToLower(expected)))
}

// Sign 生成 HMAC 签名。
func Sign(secret, timestamp, body string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(timestamp))
	mac.Write([]byte("\n"))
	mac.Write([]byte(body))
	return hex.EncodeToString(mac.Sum(nil))
}
