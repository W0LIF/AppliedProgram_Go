package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"os"
)

func deriveKey(password string) []byte {
	sum := sha256.Sum256([]byte(password))
	return sum[:]
}

func encrypt(plaintext, password string) (string, error) {
	block, err := aes.NewCipher(deriveKey(password))
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	out := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(out), nil
}

func decrypt(encoded, password string) (string, error) {
	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("base64: %w", err)
	}
	block, err := aes.NewCipher(deriveKey(password))
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	ns := gcm.NonceSize()
	if len(raw) < ns {
		return "", fmt.Errorf("ciphertext too short")
	}
	nonce, ciphertext := raw[:ns], raw[ns:]
	plain, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("decrypt: %w (неверный ключ или повреждённый шифротекст)", err)
	}
	return string(plain), nil
}

func usage() {
	fmt.Fprint(os.Stderr, `aes — симметричное шифрование (AES-256-GCM)

  aes -mode=enc -text "secret" -key "mypass"
  aes -mode=dec -text "<base64>" -key "mypass"
`)
}

func main() {
	mode := flag.String("mode", "enc", "enc | dec")
	text := flag.String("text", "", "открытый текст (enc) или base64-шифротекст (dec)")
	key := flag.String("key", "", "секретный ключ (любая строка)")
	flag.Usage = usage
	flag.Parse()

	if *text == "" || *key == "" {
		usage()
		os.Exit(2)
	}

	switch *mode {
	case "enc":
		out, err := encrypt(*text, *key)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
		fmt.Println(out)
	case "dec":
		out, err := decrypt(*text, *key)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
		fmt.Println(out)
	default:
		usage()
		os.Exit(2)
	}
}
