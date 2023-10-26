package crawler

import (
	"hash/fnv"
	"strings"
)

func hash(s string) int {
	h := fnv.New32a()
	h.Write([]byte(s))
	return int(h.Sum32())
}

func removeHttpPrefix(s string) string {
	if strings.HasPrefix(s, "https://") {
		return s[len("https://"):]
	}

	if strings.HasPrefix(s, "http://") {
		return s[len("http://"):]
	}

	return s
}
