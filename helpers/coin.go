package helpers

import "math/big"

func PipValueToCoin(value string) *big.Float {
	bip := big.NewFloat(0.000000000000000001)
	val, _, err := big.ParseFloat(value, 10, 0, big.ToZero)

	if err != nil {
		CheckErr(err)
	}
	return val.Mul(val, bip)
}
