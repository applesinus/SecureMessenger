package RC6

import (
	"errors"
	"math/rand"
)

const (
	w    = 32 // Symbol size in bits
	r    = 20 // Number of rounds
	logw = 5  // log2(w)
)

type RC6 struct {
	sessionKeys []uint32
}

func NewRC6() *RC6 {
	return &RC6{}
}

func (rc6 *RC6) Expand(key []byte) ([][]byte, error) {
	if len(key) == 0 {
		return nil, errors.New("key cannot be empty")
	}

	L := make([]uint32, (len(key)+3)/4)
	for i := 0; i < len(L); i++ {
		L[i] = 0
		for j := 0; j < 4; j++ {
			if i*4+j < len(key) {
				L[i] |= uint32(key[i*4+j]) << (8 * j)
			}
		}
	}

	S := make([]uint32, 2*r+4)
	S[0] = 0xB7E15163
	for i := 1; i < len(S); i++ {
		S[i] = S[i-1] + 0x9E3779B9
	}

	var A, B uint32
	i, j := 0, 0
	for k := 0; k < 3*len(S); k++ {
		S[i] = rotl32(S[i]+A+B, 3)
		A = S[i]
		L[j] = rotl32(L[j]+A+B, int(A+B))
		B = L[j]
		i = (i + 1) % len(S)
		j = (j + 1) % len(L)
	}

	rc6.sessionKeys = S

	roundKeys := make([][]byte, len(S))
	for i, val := range S {
		roundKeys[i] = fromUint32(val)
	}

	return roundKeys, nil
}

func (rc6 *RC6) SymmetricEncrypt(inputBlock []byte) ([]byte, error) {
	if len(inputBlock) != 16 {
		return nil, errors.New("block size must be 16 bytes")
	}

	A := toUint32(inputBlock[0:4])
	B := toUint32(inputBlock[4:8])
	C := toUint32(inputBlock[8:12])
	D := toUint32(inputBlock[12:16])

	B += rc6.sessionKeys[0]
	D += rc6.sessionKeys[1]

	for i := 1; i <= r; i++ {
		T := rotl32(B*(2*B+1), logw)
		U := rotl32(D*(2*D+1), logw)
		A = rotl32(A^T, int(U)) + rc6.sessionKeys[2*i]
		C = rotl32(C^U, int(T)) + rc6.sessionKeys[2*i+1]
		A, B, C, D = B, C, D, A
	}

	A += rc6.sessionKeys[2*r+2]
	C += rc6.sessionKeys[2*r+3]

	outputBlock := append(fromUint32(A), fromUint32(B)...)
	outputBlock = append(outputBlock, fromUint32(C)...)
	outputBlock = append(outputBlock, fromUint32(D)...)

	return outputBlock, nil
}

func (rc6 *RC6) SymmetricDecrypt(inputBlock []byte) ([]byte, error) {
	if len(inputBlock) != 16 {
		return nil, errors.New("block size must be 16 bytes")
	}

	A := toUint32(inputBlock[0:4])
	B := toUint32(inputBlock[4:8])
	C := toUint32(inputBlock[8:12])
	D := toUint32(inputBlock[12:16])

	C -= rc6.sessionKeys[2*r+3]
	A -= rc6.sessionKeys[2*r+2]

	for i := r; i >= 1; i-- {
		A, B, C, D = D, A, B, C
		T := rotl32(B*(2*B+1), logw)
		U := rotl32(D*(2*D+1), logw)
		C = rotr32(C-rc6.sessionKeys[2*i+1], int(T)) ^ U
		A = rotr32(A-rc6.sessionKeys[2*i], int(U)) ^ T
	}

	D -= rc6.sessionKeys[1]
	B -= rc6.sessionKeys[0]

	outputBlock := append(fromUint32(A), fromUint32(B)...)
	outputBlock = append(outputBlock, fromUint32(C)...)
	outputBlock = append(outputBlock, fromUint32(D)...)

	return outputBlock, nil
}

// Helper functions.
func toUint32(b []byte) uint32 {
	return uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24
}

func fromUint32(n uint32) []byte {
	return []byte{
		byte(n), byte(n >> 8), byte(n >> 16), byte(n >> 24),
	}
}

func rotl32(x uint32, y int) uint32 {
	return (x << (y % 32)) | (x >> (32 - (y % 32)))
}

func rotr32(x uint32, y int) uint32 {
	return (x >> (y % 32)) | (x << (32 - (y % 32)))
}

func GenerateIV(blockSize int) ([]byte, error) {
	iv := make([]byte, blockSize)

	for i := range iv {
		iv[i] = byte(rand.Intn(256))
	}
	return iv, nil
}
