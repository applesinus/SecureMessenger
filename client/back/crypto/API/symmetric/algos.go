package cryptocontext

import (
	"crypto/rand"
	"errors"
	"fmt"
)

const (
	blockSizeConst = 16
)

// Encryptions

func (bcp *BlockCipherProcessor) encryptECB(input []byte) ([]byte, error) {
	padded := bcp.padder.ApplyPadding(input, blockSizeConst, bcp.cipher.padding)
	//log.Printf("Padded: %s", padded)
	blockSize := blockSizeConst
	encrypted := make([]byte, len(padded))

	blocks := make(map[int]struct {
		ch    chan []byte
		block []byte
	})

	for bs := 0; bs < len(padded); bs += blockSize {
		ch := make(chan []byte)
		blocks[bs] = struct {
			ch    chan []byte
			block []byte
		}{
			ch:    ch,
			block: make([]byte, blockSize),
		}

		go func(bs int) {
			end := min(bs+blockSize, len(padded))
			encrypdedBlock, err := bcp.cipher.encryptionObject.SymmetricEncrypt(padded[bs:end])
			if err != nil {
				ch <- nil
			}
			ch <- encrypdedBlock
		}(bs)
	}

	for bs := 0; bs < len(padded); bs += blockSize {
		encryptedBlock := <-blocks[bs].ch
		if encryptedBlock == nil {
			return nil, fmt.Errorf("failed to encrypt block: %v", blocks[bs].block)
		}
		end := min(bs+blockSize, len(padded))
		copy(encrypted[bs:end], encryptedBlock)
	}

	return encrypted, nil
}

func (bcp *BlockCipherProcessor) encryptCBC(input []byte, iv []byte) ([]byte, error) {
	if len(iv) != blockSizeConst {
		return nil, fmt.Errorf("invalid IV length, got %v, expected %v", len(iv), blockSizeConst)
	}

	padded := bcp.padder.ApplyPadding(input, blockSizeConst, bcp.cipher.padding)
	//log.Printf("Padded: %s", padded)
	blockSize := blockSizeConst
	encrypted := make([]byte, 0)
	currentIV := make([]byte, blockSize)
	copy(currentIV, iv)

	for bs := 0; bs < len(padded); bs += blockSize {
		for i := 0; i < blockSize; i++ {
			padded[bs+i] ^= currentIV[i]
		}

		encrypdedBlock, err := bcp.cipher.encryptionObject.SymmetricEncrypt(padded[bs : bs+blockSize])
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt block: %v", err)
		}
		encrypted = append(encrypted, encrypdedBlock...)

		copy(currentIV, encrypdedBlock)
	}

	return encrypted, nil
}

func (bcp *BlockCipherProcessor) encryptPCBC(input []byte, iv []byte) ([]byte, error) {
	if len(iv) != blockSizeConst {
		return nil, fmt.Errorf("invalid IV length, got %v, expected %v", len(iv), blockSizeConst)
	}

	padded := bcp.padder.ApplyPadding(input, blockSizeConst, bcp.cipher.padding)
	//log.Printf("Padded: %s", padded)
	blockSize := blockSizeConst
	encrypted := make([]byte, len(padded))

	prevPlainBlock := make([]byte, blockSize)
	copy(prevPlainBlock, iv)

	for bs := 0; bs < len(padded); bs += blockSize {
		currentPlainBlock := padded[bs : bs+blockSize]

		xorBlock := make([]byte, blockSize)
		for i := 0; i < blockSize; i++ {
			xorBlock[i] = currentPlainBlock[i] ^ prevPlainBlock[i]
		}

		encryptedBlock, err := bcp.cipher.encryptionObject.SymmetricEncrypt(xorBlock)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt block: %v", err)
		}

		copy(encrypted[bs:bs+blockSize], encryptedBlock)
		copy(prevPlainBlock, currentPlainBlock)
	}

	return encrypted, nil
}

func (bcp *BlockCipherProcessor) encryptCFB(input []byte, iv []byte) ([]byte, error) {
	if len(iv) != blockSizeConst {
		return nil, fmt.Errorf("invalid IV length, got %v, expected %v", len(iv), blockSizeConst)
	}

	encrypted := make([]byte, len(input))
	currentIV := make([]byte, blockSizeConst)
	copy(currentIV, iv)

	for i := 0; i < len(input); i += blockSizeConst {
		encryptedIV, err := bcp.cipher.encryptionObject.SymmetricEncrypt(currentIV)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt block: %v", err)
		}

		end := min(i+blockSizeConst, len(input))
		for j := i; j < end; j++ {
			encrypted[j] = input[j] ^ encryptedIV[j-i]
		}

		if end-i == blockSizeConst {
			copy(currentIV, encrypted[i:end])
		}
	}

	return encrypted, nil
}

func (bcp *BlockCipherProcessor) encryptOFB(input []byte, iv []byte) ([]byte, error) {
	if len(iv) != blockSizeConst {
		return nil, fmt.Errorf("invalid IV length, got %v, expected %v", len(iv), blockSizeConst)
	}

	encrypted := make([]byte, len(input))
	currentIV := make([]byte, blockSizeConst)
	copy(currentIV, iv)

	for i := 0; i < len(input); i += blockSizeConst {
		currentIV, err := bcp.cipher.encryptionObject.SymmetricEncrypt(currentIV)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt block: %v", err)
		}

		end := min(i+blockSizeConst, len(input))
		for j := i; j < end; j++ {
			encrypted[j] = input[j] ^ currentIV[j-i]
		}
	}

	return encrypted, nil
}

func (bcp *BlockCipherProcessor) encryptCTR(input []byte, iv []byte) ([]byte, error) {
	if len(iv) != blockSizeConst {
		return nil, fmt.Errorf("invalid IV length, got %v, expected %v", len(iv), blockSizeConst)
	}

	encrypted := make([]byte, len(input))
	counter := make([]byte, blockSizeConst)
	copy(counter, iv)

	for i := 0; i < len(input); i += blockSizeConst {
		keyStream, err := bcp.cipher.encryptionObject.SymmetricEncrypt(counter)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt block: %v", err)
		}

		end := min(i+blockSizeConst, len(input))
		for j := i; j < end; j++ {
			encrypted[j] = input[j] ^ keyStream[j-i]
		}

		incrementCounter(counter)
	}

	return encrypted, nil
}

func (bcp *BlockCipherProcessor) encryptRandomDelta(input []byte, iv []byte) ([]byte, error) {
	if len(iv) != blockSizeConst {
		return nil, fmt.Errorf("invalid IV length, got %v, expected %v", len(iv), blockSizeConst)
	}

	padded := bcp.padder.ApplyPadding(input, blockSizeConst, bcp.cipher.padding)
	//log.Printf("Padded: %s", padded)
	blockSize := blockSizeConst

	encrypted := make([]byte, len(padded)+len(padded)/blockSize*blockSize)

	currentDelta := make([]byte, blockSize)
	copy(currentDelta, iv)

	for bs, out := 0, 0; bs < len(padded); bs, out = bs+blockSize, out+2*blockSize {
		currentBlock := padded[bs : bs+blockSize]

		randomDelta := make([]byte, blockSize)
		_, err := rand.Read(randomDelta)
		if err != nil {
			return nil, fmt.Errorf("failed to generate random delta: %v", err)
		}

		xorBlock := make([]byte, blockSize)
		for i := 0; i < blockSize; i++ {
			xorBlock[i] = currentBlock[i] ^ randomDelta[i]
		}

		encryptedBlock, err := bcp.cipher.encryptionObject.SymmetricEncrypt(xorBlock)
		if err != nil {
			return nil, fmt.Errorf("error encrypting block: %v", err)
		}

		copy(encrypted[out:out+blockSize], encryptedBlock)
		copy(encrypted[out+blockSize:out+2*blockSize], randomDelta)
	}

	return encrypted, nil
}

// Decryptions

func (bcp *BlockCipherProcessor) decryptECB(input []byte) ([]byte, error) {
	blockSize := blockSizeConst
	if len(input)%blockSize != 0 {
		return nil, errors.New("input length must be multiple of block size")
	}
	decrypted := make([]byte, len(input))

	blocks := make(map[int]struct {
		ch    chan []byte
		block []byte
	})

	for bs := 0; bs < len(input); bs += blockSize {
		ch := make(chan []byte)
		blocks[bs] = struct {
			ch    chan []byte
			block []byte
		}{
			ch:    ch,
			block: make([]byte, blockSize),
		}

		go func(bs int) {
			end := min(bs+blockSize, len(input))
			encrypdedBlock, err := bcp.cipher.encryptionObject.SymmetricDecrypt(input[bs:end])
			if err != nil {
				ch <- nil
			}
			ch <- encrypdedBlock
		}(bs)
	}

	for bs := 0; bs < len(input); bs += blockSize {
		encryptedBlock := <-blocks[bs].ch
		if encryptedBlock == nil {
			return nil, errors.New("failed to decrypt block")
		}
		end := min(bs+blockSize, len(input))
		copy(decrypted[bs:end], encryptedBlock)
	}

	return decrypted, nil
}

func (bcp *BlockCipherProcessor) decryptCBC(input []byte, iv []byte) ([]byte, error) {
	if len(iv) != blockSizeConst {
		return nil, fmt.Errorf("invalid IV length, got %v, expected %v", len(iv), blockSizeConst)
	}

	blockSize := blockSizeConst
	if blockSizeConst%blockSize != 0 {
		return nil, errors.New("input length must be multiple of block size")
	}

	decrypted := make([]byte, len(input))
	currentIV := make([]byte, blockSize)
	copy(currentIV, iv)

	for bs := 0; bs < len(input); bs += blockSize {
		decBlock, err := bcp.cipher.encryptionObject.SymmetricDecrypt(input[bs : bs+blockSize])
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt block: %v", err)
		}

		for i := 0; i < blockSize; i++ {
			decrypted[bs+i] = decBlock[i] ^ currentIV[i]
		}

		copy(currentIV, input[bs:bs+blockSize])
	}

	return decrypted, nil
}

func (bcp *BlockCipherProcessor) decryptPCBC(input []byte, iv []byte) ([]byte, error) {
	if len(iv) != blockSizeConst {
		return nil, fmt.Errorf("invalid IV length, got %v, expected %v", len(iv), blockSizeConst)
	}

	blockSize := blockSizeConst
	if len(input)%blockSize != 0 {
		return nil, errors.New("input length must be a multiple of block size")
	}

	decrypted := make([]byte, len(input))

	prevPlainBlock := make([]byte, blockSize)
	copy(prevPlainBlock, iv)

	for bs := 0; bs < len(input); bs += blockSize {
		currentCipherBlock := input[bs : bs+blockSize]

		decryptedBlock, err := bcp.cipher.encryptionObject.SymmetricDecrypt(currentCipherBlock)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt block: %v", err)
		}

		for i := 0; i < blockSize; i++ {
			decrypted[bs+i] = decryptedBlock[i] ^ prevPlainBlock[i]
		}

		copy(prevPlainBlock, decrypted[bs:bs+blockSize])
	}

	return decrypted, nil
}

func (bcp *BlockCipherProcessor) decryptCFB(input []byte, iv []byte) ([]byte, error) {
	if len(iv) != blockSizeConst {
		return nil, fmt.Errorf("invalid IV length, got %v, expected %v", len(iv), blockSizeConst)
	}

	decrypted := make([]byte, len(input))
	currentIV := make([]byte, blockSizeConst)
	copy(currentIV, iv)

	for i := 0; i < len(input); i += blockSizeConst {
		encryptedIV, err := bcp.cipher.encryptionObject.SymmetricEncrypt(currentIV)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt block: %v", err)
		}

		end := min(i+blockSizeConst, len(input))
		for j := i; j < end; j++ {
			decrypted[j] = input[j] ^ encryptedIV[j-i]
		}

		if end-i == blockSizeConst {
			copy(currentIV, input[i:end])
		}
	}

	return decrypted, nil
}

func (bcp *BlockCipherProcessor) decryptOFB(input []byte, iv []byte) ([]byte, error) {
	return bcp.encryptOFB(input, iv)
}

func (bcp *BlockCipherProcessor) decryptCTR(input []byte, iv []byte) ([]byte, error) {
	return bcp.encryptCTR(input, iv)
}

func (bcp *BlockCipherProcessor) decryptRandomDelta(input []byte, iv []byte) ([]byte, error) {
	if len(iv) != blockSizeConst {
		return nil, fmt.Errorf("invalid IV length, got %v, expected %v", len(iv), blockSizeConst)
	}

	blockSize := blockSizeConst
	if len(input)%(2*blockSize) != 0 {
		return nil, errors.New("input length must be multiple of 2 * block size")
	}

	decrypted := make([]byte, len(input)/2)

	currentDelta := make([]byte, blockSize)
	copy(currentDelta, iv)

	for bs, out := 0, 0; bs < len(input); bs, out = bs+2*blockSize, out+blockSize {
		currentBlock := input[bs : bs+blockSize]
		randomDelta := input[bs+blockSize : bs+2*blockSize]

		decryptedBlock, err := bcp.cipher.encryptionObject.SymmetricDecrypt(currentBlock)
		if err != nil {
			return nil, fmt.Errorf("error decrypting block: %v", err)
		}

		for i := 0; i < blockSize; i++ {
			decrypted[out+i] = decryptedBlock[i] ^ randomDelta[i]
		}
	}

	return decrypted, nil
}

// helper functions
func incrementCounter(counter []byte) {
	for i := len(counter) - 1; i >= 0; i-- {
		counter[i]++
		if counter[i] != 0 {
			break
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
