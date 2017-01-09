package utils

import (
	"crypto/rand"
	"encoding/base64"
)

func GenerateRandomBytes(length int) ([]byte, error) {
	/*
		returns securely generated random bytes.
		It will return an error if the system's secure random
		number generator fails to function correctly, in which
		case the caller should not continue.
	*/
	b := make([]byte, length)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return nil, err
	}
	return b, nil
}

func GenerateRandomBase64String(maxLength int) (string, error) {
	/*
		Returns a URL-safe, base64 encoded securely generated random string.
		It will return an error if the system's secure random number generator
		fails to function correctly, in which case the caller should not
		continue.
		The actual length of string will be `floor(maxLength / 4) * 4`
	*/
	b, err := GenerateRandomBytes(int(maxLength/4) * 3)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
