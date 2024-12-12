package cryptoType

import "strings"

const (
	ECB = iota
	CBC
	PCBC
	CFB
	OFB
	CTR
	RandomDelta
)

func GetEncryptionMode(mode string) int {
	mode = strings.ToLower(mode)

	switch mode {
	case "ecb":
		return ECB
	case "cbc":
		return CBC
	case "pcbc":
		return PCBC
	case "cfb":
		return CFB
	case "ofb":
		return OFB
	case "ctr":
		return CTR
	case "randomdelta":
		return RandomDelta
	default:
		return -1
	}
}
