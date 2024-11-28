package task1

import "log"

func ShuffleIPtest(input []byte, bitOrderIsGrowing bool, startBit int) []byte {
	log.Printf("ShuffleIP, len:%d\n", len(input))
	out := make([]byte, 0, (len(input)+7)/8*8)

	for i := 0; i < (len(input)+7)/8; i++ {
		log.Printf("ShuffleIP, iter %d\n", i)
		block := make([]byte, 8)
		if (i+1)*8 < len(input) {
			copy(block, input[i*8:(i+1)*8])
		} else {
			copy(block, input[i*8:])
			for j := 0; j < 8-(len(input)-i*8); j++ {
				block = append(block, 0)
			}
		}

		block, err := ShuffleBits(block, ip, bitOrderIsGrowing, startBit)
		if err != nil {
			panic(err)
		}
		out = append(out, block...)
	}
	log.Printf("ShuffleIP, out len:%d\n", len(out))
	return out
}

func ShuffleIPRevtest(input []byte, bitOrderIsGrowing bool, startBit int) []byte {
	out := make([]byte, len(input))
	for i := 0; i < (len(input)+7)/8; i++ {
		block := make([]byte, 8)
		if (i+1)*8 < len(input) {
			copy(block, input[i*8:(i+1)*8])
		} else {
			copy(block, input[i*8:])
			for j := 0; j < 8-(len(input)-i*8); j++ {
				block = append(block, 0)
			}
		}

		block, err := ShuffleBits(block, ipRev, bitOrderIsGrowing, startBit)
		if err != nil {
			panic(err)
		}
		out = append(out, block...)
	}
	return out
}
