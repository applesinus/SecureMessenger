package diffiehellman

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

func GeneratePrime(bits int) (*big.Int, error) {
	return rand.Prime(rand.Reader, bits)
}

func GeneratePrimitiveRoot(p *big.Int) (*big.Int, error) {
	if !p.ProbablyPrime(20) {
		return nil, fmt.Errorf("input must be a prime number")
	}

	pMinus1 := new(big.Int).Sub(p, big.NewInt(1))
	primeFactors, err := findPrimeFactors(pMinus1)
	if err != nil {
		return nil, err
	}

	for {
		candidate, err := rand.Int(rand.Reader, pMinus1)
		if err != nil {
			return nil, err
		}
		candidate.Add(candidate, big.NewInt(2))

		if isPrimitiveRoot(candidate, p, primeFactors) {
			return candidate, nil
		}
	}
}

func findPrimeFactors(n *big.Int) ([]*big.Int, error) {
	factors := []*big.Int{}
	two := big.NewInt(2)

	num := new(big.Int).Set(n)

	for new(big.Int).Mod(num, two).Cmp(big.NewInt(0)) == 0 {
		if len(factors) == 0 || factors[len(factors)-1].Cmp(two) != 0 {
			factors = append(factors, new(big.Int).Set(two))
		}
		num.Div(num, two)
	}

	for f := big.NewInt(3); f.Mul(f, f).Cmp(num) <= 0; f.Add(f, two) {
		for new(big.Int).Mod(num, f).Cmp(big.NewInt(0)) == 0 {
			if len(factors) == 0 || factors[len(factors)-1].Cmp(f) != 0 {
				factors = append(factors, new(big.Int).Set(f))
			}
			num.Div(num, f)
		}
	}

	if num.Cmp(big.NewInt(1)) > 0 {
		factors = append(factors, num)
	}

	return factors, nil
}

func isPrimitiveRoot(g, p *big.Int, primeFactors []*big.Int) bool {
	if g.Cmp(big.NewInt(2)) < 0 || g.Cmp(new(big.Int).Sub(p, big.NewInt(1))) >= 0 {
		return false
	}

	pMinus1 := new(big.Int).Sub(p, big.NewInt(1))
	for _, factor := range primeFactors {
		exponent := new(big.Int).Div(pMinus1, factor)
		check := new(big.Int).Exp(g, exponent, p)

		if check.Cmp(big.NewInt(1)) == 0 {
			return false
		}
	}

	return true
}

func GeneratePrivateKey(p *big.Int) (*big.Int, error) {
	return rand.Int(rand.Reader, new(big.Int).Sub(p, big.NewInt(2)))
}

func GeneratePublicKey(privateKey, g, p *big.Int) *big.Int {
	return new(big.Int).Exp(g, privateKey, p)
}

func ComputeSharedSecret(privateKey, otherPublicKey, p *big.Int) *big.Int {
	return new(big.Int).Exp(otherPublicKey, privateKey, p)
}
