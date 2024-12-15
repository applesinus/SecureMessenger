package cryptocontext

import (
	"crypto/rand"
	"fmt"
	"messengerClient/back/crypto/constants/cryptoType"
	magenta "messengerClient/back/crypto/tasks/MAGENTA"
	"messengerClient/back/crypto/tasks/RC6"
	"strings"
)

type SymmetricContext struct {
	mode             int
	padding          int
	encryptionObject SymmetricEncryptionInterface
	key              []byte
	iv               []byte
	additionalParams []interface{}
}

func NewSymmetricContext(
	key []byte,
	mode int,
	padding int,
	encryptionImpl SymmetricEncryptionInterface,
	iv []byte,
	additionalParams ...interface{},
) (*SymmetricContext, error) {
	ctx := &SymmetricContext{
		mode:             mode,
		padding:          padding,
		encryptionObject: encryptionImpl,
		additionalParams: additionalParams,
		key:              key,
	}

	_, err := ctx.encryptionObject.Expand(key)
	if err != nil {
		return nil, err
	}

	if ctx.mode != cryptoType.ECB && iv == nil {
		ctx.iv = GenerateIV()
	}

	if ctx.mode == cryptoType.ECB && iv != nil {
		ctx.iv = nil
	}

	return ctx, nil
}

func (ctx *SymmetricContext) Encrypt(input []byte) ([]byte, []byte, error) {
	bcp := &BlockCipherProcessor{padder: PaddingHelper{}, cipher: *ctx}
	encrypted, err := bcp.Encrypt(input)
	if err != nil {
		return nil, nil, err
	}
	return ctx.iv, encrypted, nil
}

func (ctx *SymmetricContext) Decrypt(input []byte) ([]byte, error) {
	bcp := &BlockCipherProcessor{padder: PaddingHelper{}, cipher: *ctx}
	return bcp.Decrypt(input)
}

type SymmetricEncryptionInterface interface {
	Expand(key []byte) ([][]byte, error)
	SymmetricEncrypt(input []byte) ([]byte, error)
	SymmetricDecrypt(input []byte) ([]byte, error)
}

func GetSymmetricMode(encryption string) SymmetricEncryptionInterface {
	encryption = strings.ToLower(encryption)

	switch encryption {
	case "rc6":
		return RC6.NewRC6()
	case "magenta":
		return magenta.NewMagenta()
	default:
		return nil
	}
}

type BlockCipherProcessor struct {
	padder PaddingHelper
	cipher SymmetricContext
}

func (bcp *BlockCipherProcessor) Encrypt(
	input []byte,
) ([]byte, error) {
	switch bcp.cipher.mode {
	case cryptoType.ECB:
		return bcp.encryptECB(input)
	case cryptoType.CBC:
		return bcp.encryptCBC(input, bcp.cipher.iv)
	case cryptoType.PCBC:
		return bcp.encryptPCBC(input, bcp.cipher.iv)
	case cryptoType.CFB:
		return bcp.encryptCFB(input, bcp.cipher.iv)
	case cryptoType.OFB:
		return bcp.encryptOFB(input, bcp.cipher.iv)
	case cryptoType.CTR:
		return bcp.encryptCTR(input, bcp.cipher.iv)
	case cryptoType.RandomDelta:
		return bcp.encryptRandomDelta(input, bcp.cipher.iv)
	default:
		return nil, fmt.Errorf("unsupported encryption mode: %d", bcp.cipher.mode)
	}
}

func (bcp *BlockCipherProcessor) Decrypt(
	input []byte,
) ([]byte, error) {
	var decrypted []byte
	var err error

	switch bcp.cipher.mode {
	case cryptoType.ECB:
		decrypted, err = bcp.decryptECB(input)
	case cryptoType.CBC:
		decrypted, err = bcp.decryptCBC(input, bcp.cipher.iv)
	case cryptoType.PCBC:
		decrypted, err = bcp.decryptPCBC(input, bcp.cipher.iv)
	case cryptoType.CFB:
		decrypted, err = bcp.decryptCFB(input, bcp.cipher.iv)
	case cryptoType.OFB:
		decrypted, err = bcp.decryptOFB(input, bcp.cipher.iv)
	case cryptoType.CTR:
		decrypted, err = bcp.decryptCTR(input, bcp.cipher.iv)
	case cryptoType.RandomDelta:
		decrypted, err = bcp.decryptRandomDelta(input, bcp.cipher.iv)
	default:
		return nil, fmt.Errorf("unsupported decryption mode: %d", bcp.cipher.mode)
	}

	if err != nil {
		return nil, err
	}

	//log.Printf("Decrypted: %s", decrypted)

	if bcp.cipher.mode == cryptoType.CFB || bcp.cipher.mode == cryptoType.OFB || bcp.cipher.mode == cryptoType.CTR {
		return decrypted, nil
	}

	return bcp.padder.RemovePadding(decrypted, bcp.cipher.padding), nil
}

func GenerateIV() []byte {
	iv := make([]byte, 16)

	_, err := rand.Read(iv)
	if err != nil {
		panic("Ошибка при генерации IV: " + err.Error())
	}

	return iv
}
