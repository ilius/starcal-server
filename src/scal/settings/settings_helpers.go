package settings

import (
	"bufio"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
)

var log = NewLogger("settings")
var stdin = bufio.NewReader(os.Stdin)
var masterKey []byte

func getMasterKey() []byte {
	var err error
	if masterKey != nil {
		return masterKey
	}
	valueStr := os.Getenv("MASTER_KEY")
	if valueStr != "" {
		value, err := hex.DecodeString(valueStr)
		if err != nil {
			panic(fmt.Errorf("MASTER_KEY must be hex-encoded: %v", err))
		}
		masterKey = value
		return masterKey
	}
	for i := 0; i < 10; i++ {
		valueStr, err = stdin.ReadString('\n')
		if err != nil {
			log.Error(err)
			continue
		}
		valueStr = strings.TrimRight(valueStr, "\n")
		value, err := hex.DecodeString(valueStr)
		if err != nil {
			log.Error("Master key must be hex-encoded")
			continue
		}
		masterKey = value
		break
	}
	if masterKey == nil {
		panic("Failed to read master key from environment or stdin")
	}
	return masterKey
}

func decryptBytesCBC(aesKey []byte, text []byte) ([]byte, error) {
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, err
	}
	if len(text) < aes.BlockSize {
		return nil, fmt.Errorf("ciphertext too short")
	}
	iv := text[:aes.BlockSize]
	plaintext := make([]byte, len(text[aes.BlockSize:]))
	copy(plaintext, text[aes.BlockSize:])
	cfb := cipher.NewCBCDecrypter(block, iv)
	cfb.CryptBlocks(plaintext, plaintext)
	return plaintext, nil
}

func secretCBC(valueEncBase64 string) string {
	aesKey := getMasterKey()
	valueEnc, err := base64.StdEncoding.DecodeString(valueEncBase64)
	if err != nil {
		panic(fmt.Errorf("bad base64-encoded secret: %v", err))
	}
	valueB, err := decryptBytesCBC(aesKey, valueEnc)
	if err != nil {
		panic(fmt.Errorf("bad ecrypted secret: %v", err))
	}
	valueB = bytes.TrimRight(valueB, "\x00")
	return string(valueB)
}
