package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"
)

func TokenHash(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func AccessTokenKey(token string) string { return AccessTokenHashKey(TokenHash(token)) }

func AccessTokenHashKey(hash string) string { return "auth:access:" + hash }

func RefreshTokenKey(token string) string { return "auth:refresh:" + TokenHash(token) }

func UserSessionsKey(userID int64, clientID string) string {
	return "auth:user-sessions:" + strconv.FormatInt(userID, 10) + ":" + clientID
}
