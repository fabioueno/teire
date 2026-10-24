package strategy

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/tidwall/gjson"
)

// These characters start gjson queries (wildcards, arrays, modifiers,
// pipes), which sjson cannot write.
const querySyntax = "*?#@|!"

func checkPath(path string) error {
	if path == "" {
		return errors.New("owned path is empty")
	}

	if strings.ContainsAny(path, querySyntax) {
		return fmt.Errorf("owned path %q uses query syntax; use plain keys separated by dots", path)
	}

	return nil
}

func hashValue(r gjson.Result) string {
	if !r.Exists() {
		return ""
	}

	return hashRaw(r.Raw)
}

func hashRaw(raw string) string {
	sum := sha256.Sum256(compact(raw))
	return hex.EncodeToString(sum[:])
}

// compact drops insignificant whitespace, so a value reformatted by another
// program still hashes the same.
func compact(raw string) []byte {
	var buf bytes.Buffer
	if err := json.Compact(&buf, []byte(raw)); err != nil {
		return []byte(raw)
	}

	return buf.Bytes()
}
