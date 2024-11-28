package internal

/*import (
	"infosec/constants/cryptoType"
	"infosec/constants/paddingType"
)*/

type KeyExpander interface {
	ExpandKey(key []byte) error
}

type Encryptor interface {
	Encrypt(block []byte, roundKeys [][]byte) ([]byte, error)
}

type Decryptor interface {
	Decrypt(block []byte, roundKeys [][]byte) ([]byte, error)
}

type Cypherer interface {
	KeyExpander
	Encryptor
	Decryptor
}

type Task1Cypherer struct {
	cryptoType  int
	paddingType int
	roundKeys   [][]byte
}

func NewTask1Cypherer(key []byte, cryptoType int, paddingType int, initVect []byte, options map[string]any) (*Task1Cypherer, error) {
	t1c := Task1Cypherer{
		cryptoType:  cryptoType,
		paddingType: paddingType,
	}

	err := t1c.ExpandKey(key)
	if err != nil {
		return nil, err
	}

	return &t1c, nil
}

func (t1c *Task1Cypherer) ExpandKey(key []byte) error {
	// TODO
	return nil
}

func (t1c *Task1Cypherer) Encrypt(block []byte, roundKeys [][]byte) ([]byte, error) {
	// TODO
	return nil, nil
}

func (t1c *Task1Cypherer) Decrypt(block []byte, roundKeys [][]byte) ([]byte, error) {
	// TODO
	return nil, nil
}
