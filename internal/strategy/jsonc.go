package strategy

import (
	"bytes"
	"encoding/json"
	"slices"
	"strings"

	"github.com/tailscale/hujson"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// Strict JSON goes through gjson and sjson, which keep every other byte. A
// file with comments or trailing commas, as editors keep their settings, is
// not strict JSON and goes through hujson, which keeps them too.

func validDocument(data []byte) bool {
	if gjson.ValidBytes(data) {
		return true
	}

	_, err := hujson.Parse(data)

	return err == nil
}

// lookup gives the value at an owned path as plain JSON, so comments and
// trailing commas never reach a hash.
func lookup(data []byte, path string) gjson.Result {
	if gjson.ValidBytes(data) {
		return gjson.GetBytes(data, path)
	}

	root, err := hujson.Parse(data)
	if err != nil {
		return gjson.Result{}
	}

	found := root.Find(pointer(splitPath(path)))
	if found == nil {
		return gjson.Result{}
	}

	plain, err := hujson.Standardize(found.Pack())
	if err != nil {
		return gjson.Result{}
	}

	return gjson.ParseBytes(bytes.TrimSpace(plain))
}

func setPath(data []byte, path string, value []byte) ([]byte, error) {
	if gjson.ValidBytes(data) {
		return sjson.SetRawBytes(data, path, value)
	}

	return setJSONC(data, splitPath(path), value)
}

func deletePath(data []byte, path string) ([]byte, error) {
	if gjson.ValidBytes(data) {
		return sjson.DeleteBytes(data, path)
	}

	return deleteJSONC(data, splitPath(path))
}

type patchOp struct {
	Op    string          `json:"op"`
	Path  string          `json:"path"`
	Value json.RawMessage `json:"value,omitempty"`
}

// setJSONC adds the value at the path. A missing parent object is created,
// as sjson does, by adding the value wrapped in the keys that are missing.
func setJSONC(data []byte, keys []string, value []byte) ([]byte, error) {
	root, err := hujson.Parse(data)
	if err != nil {
		return nil, err
	}

	have := len(keys)
	for have > 0 && root.Find(pointer(keys[:have])) == nil {
		have--
	}

	at := keys
	if have < len(keys) {
		at = keys[:have+1]

		value, err = wrap(keys[have+1:], value)
		if err != nil {
			return nil, err
		}
	}

	if err := patch(&root, patchOp{Op: "add", Path: pointer(at), Value: value}); err != nil {
		return nil, err
	}

	return root.Pack(), nil
}

// deleteJSONC leaves the file alone when the path is already absent, as
// sjson does.
func deleteJSONC(data []byte, keys []string) ([]byte, error) {
	root, err := hujson.Parse(data)
	if err != nil {
		return nil, err
	}

	if root.Find(pointer(keys)) == nil {
		return data, nil
	}

	if err := patch(&root, patchOp{Op: "remove", Path: pointer(keys)}); err != nil {
		return nil, err
	}

	return root.Pack(), nil
}

func patch(root *hujson.Value, op patchOp) error {
	ops, err := json.Marshal([]patchOp{op})
	if err != nil {
		return err
	}

	return root.Patch(ops)
}

func wrap(keys []string, value []byte) ([]byte, error) {
	for _, key := range slices.Backward(keys) {
		var err error

		value, err = json.Marshal(map[string]json.RawMessage{key: value})
		if err != nil {
			return nil, err
		}
	}

	return value, nil
}

// splitPath cuts an owned path at its unescaped dots. A backslash makes the
// next character a plain one, as in gjson.
func splitPath(path string) []string {
	var keys []string

	var key strings.Builder

	escaped := false

	for _, r := range path {
		switch {
		case escaped:
			key.WriteRune(r)

			escaped = false
		case r == '\\':
			escaped = true
		case r == '.':
			keys = append(keys, key.String())
			key.Reset()
		default:
			key.WriteRune(r)
		}
	}

	return append(keys, key.String())
}

// pointer makes an RFC 6901 JSON Pointer, which escapes "~" and "/" in a key.
func pointer(keys []string) string {
	escape := strings.NewReplacer("~", "~0", "/", "~1")

	var b strings.Builder
	for _, key := range keys {
		b.WriteString("/" + escape.Replace(key))
	}

	return b.String()
}
