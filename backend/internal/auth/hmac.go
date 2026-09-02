package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
)

func GenerateHMAC(payload map[string]interface{}, secret string) string {
	sortedKeys := make([]string, 0, len(payload))
	for k := range payload {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)

	var sb strings.Builder
	for i, k := range sortedKeys {
		if i > 0 {
			sb.WriteString("&")
		}
		sb.WriteString(fmt.Sprintf("%s=%v", k, payload[k]))
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(sb.String()))
	return hex.EncodeToString(mac.Sum(nil))
}

func VerifyHMAC(payload map[string]interface{}, signature string, secret string) bool {
	expected := GenerateHMAC(payload, secret)
	return hmac.Equal([]byte(expected), []byte(signature))
}
