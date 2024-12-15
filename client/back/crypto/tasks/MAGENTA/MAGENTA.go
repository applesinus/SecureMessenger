package magenta

import (
	"errors"
	"fmt"
)

type SymmetricEncryptionInterface interface {
	Expand(key []byte) ([][]byte, error)
	SymmetricEncrypt(input []byte) ([]byte, error)
	SymmetricDecrypt(input []byte) ([]byte, error)
}

type MAGENTA struct {
	roundKeys [][]byte
	nRounds   int
}

func NewMagenta() *MAGENTA {
	return &MAGENTA{}
}

func (m *MAGENTA) Expand(key []byte) ([][]byte, error) {
	var roundKeys [][]byte
	switch len(key) {
	case 16:
		m.nRounds = 6
		roundKeys = make([][]byte, 6*2)
		m.roundKeys = roundKeys
		k1 := key[:8]
		k2 := key[8:]

		for i := 0; i < 6; i++ {
			m.roundKeys[i*2] = k1
			m.roundKeys[i*2+1] = k2

			k1 = append(k1[1:], k1[0])
			k2 = append(k2[1:], k2[0])
		}
	case 24:
		m.nRounds = 6
		roundKeys = make([][]byte, 6*3)
		m.roundKeys = roundKeys
		k1, k2, k3 := key[:8], key[8:16], key[16:]

		for i := 0; i < 6; i++ {
			m.roundKeys[i*3] = k1
			m.roundKeys[i*3+1] = k2
			m.roundKeys[i*3+2] = k3

			k1 = append(k1[1:], k1[0])
			k2 = append(k2[1:], k2[0])
			k3 = append(k3[1:], k3[0])
		}
	case 32:
		m.nRounds = 8
		roundKeys = make([][]byte, 8*4)
		m.roundKeys = roundKeys
		k1, k2, k3, k4 := key[:8], key[8:16], key[16:24], key[24:]

		for i := 0; i < 8; i++ {
			m.roundKeys[i*4] = k1
			m.roundKeys[i*4+1] = k2
			m.roundKeys[i*4+2] = k3
			m.roundKeys[i*4+3] = k4

			k1 = append(k1[1:], k1[0])
			k2 = append(k2[1:], k2[0])
			k3 = append(k3[1:], k3[0])
			k4 = append(k4[1:], k4[0])
		}
	default:
		return nil, fmt.Errorf("wrong key length: %d", len(key))
	}

	return roundKeys, nil
}

func (m *MAGENTA) SymmetricEncrypt(input []byte) ([]byte, error) {
	if len(input)%16 != 0 {
		return nil, errors.New("length of input must be a multiple of 16")
	}

	output := make([]byte, len(input))
	for i := 0; i < len(input); i += 16 {
		x1 := input[i : i+8]
		x2 := input[i+8 : i+16]
		x1Enc, x2Enc := m.encryptBlock(x1, x2)

		copy(output[i:i+8], x1Enc)
		copy(output[i+8:i+16], x2Enc)
	}

	return output, nil
}

func (m *MAGENTA) SymmetricDecrypt(input []byte) ([]byte, error) {
	if len(input)%16 != 0 {
		return nil, errors.New("length of input must be a multiple of 16")
	}
	output := make([]byte, len(input))

	for i := 0; i < len(input); i += 16 {
		x1 := input[i : i+8]
		x2 := input[i+8 : i+16]
		x1Dec, x2Dec := m.decryptBlock(x1, x2)

		copy(output[i:i+8], x1Dec)
		copy(output[i+8:i+16], x2Dec)
	}

	return output, nil
}

func (m *MAGENTA) encryptBlock(X1, X2 []byte) ([]byte, []byte) {
	X1Copy := append([]byte{}, X1...)
	X2Copy := append([]byte{}, X2...)

	for r := 0; r < m.nRounds; r++ {
		Kn := m.roundKeys[r]
		FResult := F(X2Copy, Kn)
		X1Copy, X2Copy = X2Copy, xorBlocks8(X1Copy, FResult)
	}

	return X1Copy, X2Copy
}

func (m *MAGENTA) decryptBlock(X1, X2 []byte) ([]byte, []byte) {
	X1Copy := append([]byte{}, X1...)
	X2Copy := append([]byte{}, X2...)
	for r := m.nRounds - 1; r >= 0; r-- {
		Kn := m.roundKeys[r]
		FResult := F(X1Copy, Kn)
		X1Copy, X2Copy = xorBlocks8(X2Copy, FResult), X1Copy
	}
	return X1Copy, X2Copy
}

func xorBlocks8(a, b []byte) []byte {
	result := make([]byte, 8)
	for i := 0; i < 8; i++ {
		result[i] = a[i] ^ b[i]
	}
	return result
}

func F(X2, Kn []byte) []byte {
	var input [16]byte
	copy(input[:8], X2)
	copy(input[8:], Kn)

	cResult := C(3, input)
	sResult := S(cResult)

	result := make([]byte, 8)
	for i := 0; i < 8; i++ {
		result[i] = sResult[2*i]
	}
	return result
}

var SBox = [256]byte{
	99, 124, 119, 123, 242, 107, 111, 197,
	48, 1, 103, 43, 254, 215, 171, 118,
	202, 130, 201, 125, 250, 89, 71, 240,
	173, 212, 162, 175, 156, 164, 114, 192,
	183, 253, 147, 38, 54, 63, 247, 204,
	52, 165, 229, 241, 113, 216, 49, 21,
	4, 199, 35, 195, 24, 150, 5, 154,
	7, 18, 128, 226, 235, 39, 178, 117,
	9, 131, 44, 26, 27, 110, 90, 160,
	82, 59, 214, 179, 41, 227, 47, 132,
	83, 209, 0, 237, 32, 252, 177, 91,
	106, 203, 190, 57, 74, 76, 88, 207,
	208, 239, 170, 251, 67, 77, 51, 133,
	69, 249, 2, 127, 80, 60, 159, 168,
	81, 163, 64, 143, 146, 157, 56, 245,
	188, 182, 218, 33, 16, 255, 243, 210,
	205, 12, 19, 236, 95, 151, 68, 23,
	196, 167, 126, 61, 100, 93, 25, 115,
	96, 129, 79, 220, 34, 42, 144, 136,
	70, 238, 184, 20, 222, 94, 11, 219,
	224, 50, 58, 10, 73, 6, 36, 92,
	194, 211, 172, 98, 145, 149, 228, 121,
	231, 200, 55, 109, 141, 213, 78, 169,
	108, 86, 244, 234, 101, 122, 174, 8,
	186, 120, 37, 46, 28, 166, 180, 198,
	232, 221, 116, 31, 75, 189, 139, 138,
	112, 62, 181, 102, 72, 3, 246, 14,
	97, 53, 87, 185, 134, 193, 29, 158,
	225, 248, 152, 17, 105, 217, 142, 148,
	155, 30, 135, 233, 206, 85, 40, 223,
	140, 161, 137, 13, 191, 230, 66, 104,
	65, 153, 45, 15, 176, 84, 187, 22,
}

func f(x byte) byte {
	return SBox[x]
}

// A(x, y) = f(x ⊕ f(y))
func A(x, y byte) byte {
	return f(x ^ f(y))
}

// PE(x, y)
func PE(x, y byte) [2]byte {
	return [2]byte{A(x, y), A(y, x)}
}

// П(X)
func П(X [16]byte) [16]byte {
	var result [16]byte
	for i := 0; i < 8; i++ {
		pe := PE(X[i], X[i+8])
		result[2*i] = pe[0]
		result[2*i+1] = pe[1]
	}
	return result
}

// T(X)
func T(X [16]byte) [16]byte {
	for i := 0; i < 4; i++ {
		X = П(X)
	}
	return X
}

// S(X)
func S(X [16]byte) [16]byte {
	var result [16]byte
	for i := 0; i < 8; i++ {
		result[i] = X[2*i]
		result[8+i] = X[2*i+1]
	}
	return result
}

// C(k, X)
func C(k int, X [16]byte) [16]byte {
	if k == 1 {
		return T(X)
	}
	prev := C(k-1, X)
	return T(xorBlocks(X, S(prev)))
}

func xorBlocks(a, b [16]byte) [16]byte {
	var result [16]byte
	for i := 0; i < 16; i++ {
		result[i] = a[i] ^ b[i]
	}
	return result
}
