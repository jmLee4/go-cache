package consistenthash

import (
	"strconv"
	"testing"
)

func TestHashing(t *testing.T) {
	hash := New(3, func(key []byte) uint32 {
		i, _ := strconv.Atoi(string(key))
		return uint32(i)
	})
	hash.InitNodes("6", "4", "2")

	testCases := map[string]string{
		"2":  "2",
		"11": "2",
		"23": "4",
		"27": "2",
	}
	for k, v := range testCases {
		if hash.GetNode(k) != v {
			t.Errorf("asking for %s, should have yielded %s", k, v)
		}
	}

	hash.InitNodes("8")
	testCases["27"] = "8"
	for k, v := range testCases {
		if hash.GetNode(k) != v {
			t.Errorf("asking for %s, should have yielded %s", k, v)
		}
	}
}
