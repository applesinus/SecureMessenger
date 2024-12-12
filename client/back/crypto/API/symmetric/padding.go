package cryptocontext

import (
	"bytes"
	"messengerClient/back/crypto/constants/paddingType"
)

type PaddingHelper struct{}

func (ph *PaddingHelper) ApplyPadding(input []byte, blockSize int, pType int) []byte {
	switch pType {
	case paddingType.Zeros:
		return ph.zerosPadding(input, blockSize)
	case paddingType.PKCS7:
		return ph.pkcs7Padding(input, blockSize)
	case paddingType.ANSIX923:
		return ph.ansix923Padding(input, blockSize)
	case paddingType.ISO10126:
		return ph.iso10126Padding(input, blockSize)
	default:
		return input
	}
}

func (ph *PaddingHelper) RemovePadding(input []byte, pType int) []byte {
	switch pType {
	case paddingType.Zeros:
		return ph.removeZerosPadding(input)
	case paddingType.PKCS7:
		return ph.removePKCS7Padding(input)
	case paddingType.ANSIX923:
		return ph.removeANSIX923Padding(input)
	case paddingType.ISO10126:
		return ph.removeISO10126Padding(input)
	default:
		return input
	}
}

func (ph *PaddingHelper) zerosPadding(input []byte, blockSize int) []byte {
	padLength := blockSize - (len(input) % blockSize)
	return append(input, bytes.Repeat([]byte{0}, padLength)...)
}

func (ph *PaddingHelper) removeZerosPadding(input []byte) []byte {
	for i := len(input) - 1; i >= 0; i-- {
		if input[i] != 0 {
			return input[:i+1]
		}
	}
	return []byte{}
}

func (ph *PaddingHelper) pkcs7Padding(input []byte, blockSize int) []byte {
	padLength := blockSize - (len(input) % blockSize)
	pad := bytes.Repeat([]byte{byte(padLength)}, padLength)
	return append(input, pad...)
}

func (ph *PaddingHelper) removePKCS7Padding(input []byte) []byte {
	padLength := int(input[len(input)-1])
	return input[:len(input)-padLength]
}

func (ph *PaddingHelper) ansix923Padding(input []byte, blockSize int) []byte {
	padLength := blockSize - (len(input) % blockSize)
	pad := make([]byte, padLength)
	pad[padLength-1] = byte(padLength)
	return append(input, pad...)
}

func (ph *PaddingHelper) removeANSIX923Padding(input []byte) []byte {
	padLength := int(input[len(input)-1])
	return input[:len(input)-padLength]
}

func (ph *PaddingHelper) iso10126Padding(input []byte, blockSize int) []byte {
	padLength := blockSize - (len(input) % blockSize)
	pad := make([]byte, padLength)
	for i := 0; i < padLength-1; i++ {
		pad[i] = byte(i % 256)
	}
	pad[padLength-1] = byte(padLength)
	return append(input, pad...)
}

func (ph *PaddingHelper) removeISO10126Padding(input []byte) []byte {
	padLength := int(input[len(input)-1])
	return input[:len(input)-padLength]
}
