package paddingType

import "strings"

const (
	Zeros = iota
	ANSIX923
	PKCS7
	ISO10126
)

func GetPaddingMode(padding string) int {
	padding = strings.ToLower(padding)

	switch padding {
	case "zeros":
		return Zeros
	case "ansix923":
		return ANSIX923
	case "pkcs7":
		return PKCS7
	case "iso10126":
		return ISO10126
	default:
		return Zeros
	}
}
