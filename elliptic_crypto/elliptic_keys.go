package elliptic_crypto

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"math/rand"
)

type Keys struct {
	publicX, publicY   int
	privateX, privateY int
	n                  int
	c                  *Curve
}

type PublicKeyInterface interface {
	GeneratePublicKey(nA int, c *Curve) (int, int)
	GeneratePrivateKey(pbX int, pbY int)
	ComparePrivateKeys(other *Keys) bool
}

func (pk *Keys) GeneratePublicKey(nA int, c *Curve) (int, int) {
	if nA >= c.M {
		return 0, 0
	}
	if nA < 2 {
		return 0, 0
	}
	pk.publicX, pk.publicY = c.ScalarOriginMultiply(nA)
	pk.n = nA
	pk.c = c
	return pk.publicX, pk.publicY
}

func (pk *Keys) GeneratePrivateKey(pbX int, pbY int) {
	pk.privateX, pk.privateY = pk.c.ScalarMultiply(pk.n, pbX, pbY)
	fmt.Println(pk.privateX, pk.privateY)
}

func (pk *Keys) ComparePrivateKeys(other *Keys) bool {
	return pk.privateX == other.privateX && pk.privateY == other.privateY
}

func (pk *Keys) SignMessage(m string) (int, int) {
	hash := sha256.New()
	hash.Write([]byte(m))
	bigC := big.NewInt(int64(pk.c.C))
	hashInt, ok := new(big.Int).SetString(hex.EncodeToString(hash.Sum(nil)), 16)
	if !ok {
		return 0, 0
	}
	bitLengthHash := uint(hashInt.BitLen())
	bitLengthC := uint(bigC.BitLen())
	hashInt = new(big.Int).Rsh(hashInt, bitLengthHash-bitLengthC)
	continueLooking := true
	var r, s int
	for continueLooking {
		randomInt := (rand.Int() + 1) % (pk.c.C - 1)
		resX, _ := pk.c.ScalarOriginMultiply(randomInt)
		r = resX % pk.c.C
		if r == 0 {
			continue
		}
		bigK := big.NewInt(int64(randomInt))
		sum := new(big.Int).Add(hashInt, big.NewInt(int64(pk.n*r)))
		sumMod := new(big.Int).Mod(sum, bigC)
		invK := new(big.Int).ModInverse(bigK, bigC)
		notMod := new(big.Int).Mul(invK, sumMod)
		s = int(new(big.Int).Mod(notMod, bigC).Int64())
		if s == 0 {
			continue
		}
		continueLooking = false
	}
	return r, s
}

func (pk *Keys) CheckSignature(m string, r, s int, paX, paY int) bool {
	hash := sha256.New()
	hash.Write([]byte(m))
	bigC := big.NewInt(int64(pk.c.C))
	hashInt, ok := new(big.Int).SetString(hex.EncodeToString(hash.Sum(nil)), 16)
	if !ok {
		return false
	}
	bitLengthHash := uint(hashInt.BitLen())
	bitLengthC := uint(bigC.BitLen())
	hashInt = new(big.Int).Rsh(hashInt, bitLengthHash-bitLengthC)
	if r < 1 || r > pk.c.C-1 || s < 1 || s > pk.c.C-1 {
		return false
	}
	bigS := big.NewInt(int64(s))
	bigR := big.NewInt(int64(r))
	w := new(big.Int).ModInverse(bigS, bigC)
	wh := new(big.Int).Mul(hashInt, w)
	u1 := new(big.Int).Mod(wh, bigC)
	rw := new(big.Int).Mul(bigR, w)
	u2 := new(big.Int).Mod(rw, bigC)
	total1X, total1Y := pk.c.ScalarOriginMultiply(int(u1.Int64()))
	total2X, total2Y := pk.c.ScalarMultiply(int(u2.Int64()), paX, paY)
	finalX, _ := pk.c.Add(total1X, total1Y, total2X, total2Y)
	if r != (finalX % pk.c.C) {
		return false
	}
	return true
}
