// +build gofuzz

package matchrelay

import (
	"github.com/coredns/coredns/plugin/pkg/fuzz"
)

// Fuzz fuzzes cache.
func Fuzz(data []byte) int {
	w := MatchRelay{}
	return fuzz.Do(w, data)
}
