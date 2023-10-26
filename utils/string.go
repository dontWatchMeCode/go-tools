package utils

import (
	"encoding/hex"
	"strings"

	"golang.org/x/crypto/sha3"
)

func Hash(s string) string {
	h := sha3.New256()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

func RemoveHttpPrefix(s string) string {
	if strings.HasPrefix(s, "https://") {
		return s[len("https://"):]
	}

	if strings.HasPrefix(s, "http://") {
		return s[len("http://"):]
	}

	return s
}
