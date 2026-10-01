package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

type faceChallengeClaims struct {
	UserID    uuid.UUID `json:"user_id"`
	Challenge string    `json:"challenge"`
	ExpiresAt int64     `json:"expires_at"`
}

func NewFaceChallenge(userID uuid.UUID, secret string) (string, string, time.Time, error) {
	random := make([]byte, 1)
	if _, err := rand.Read(random); err != nil {
		return "", "", time.Time{}, err
	}
	challenge := "turn_left"
	if random[0]&1 == 1 {
		challenge = "turn_right"
	}
	expiresAt := time.Now().Add(2 * time.Minute)
	claims, _ := json.Marshal(faceChallengeClaims{UserID: userID, Challenge: challenge, ExpiresAt: expiresAt.Unix()})
	payload := base64.RawURLEncoding.EncodeToString(claims)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	token := payload + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return challenge, token, expiresAt, nil
}

func VerifyFaceChallenge(token, challenge string, userID uuid.UUID, secret string) error {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return errors.New("invalid liveness challenge")
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(parts[0]))
	signature, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil || !hmac.Equal(signature, mac.Sum(nil)) {
		return errors.New("invalid liveness challenge")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return errors.New("invalid liveness challenge")
	}
	var claims faceChallengeClaims
	if json.Unmarshal(payload, &claims) != nil || claims.UserID != userID || claims.Challenge != challenge {
		return errors.New("liveness challenge does not match")
	}
	if time.Now().Unix() > claims.ExpiresAt {
		return errors.New("liveness challenge expired")
	}
	return nil
}
