package crypto

import (
	"crypto/md5"
	"encoding/hex"
)

func Hash(original string) string {
	hasher := md5.New()
	hasher.Write([]byte(original))
	hashed := hasher.Sum(nil)
	return hex.EncodeToString(hashed)
}
