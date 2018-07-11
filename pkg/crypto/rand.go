package crypto

import "crypto/rand"

func Rand32() ([32]byte, error) {
	var out [32]byte
	_, err := rand.Read(out[:])
	if err != nil {
		return out, err
	}

	return out, nil
}
