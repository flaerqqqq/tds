package session

import (
	"crypto/md5"
	"encoding/hex"
	"path/filepath"
	"strings"
)

// Name returns "<basename with dots as underscores>-<first 4 hex of md5(path)>".
// It must match the bash implementation so existing sessions are found.
func Name(path string) string {
	base := strings.ReplaceAll(filepath.Base(path), ".", "_")
	sum := md5.Sum([]byte(path))
	return base + "-" + hex.EncodeToString(sum[:])[:4]
}
