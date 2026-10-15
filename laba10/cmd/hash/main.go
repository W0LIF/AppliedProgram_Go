package main

import (
	"crypto/md5"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"flag"
	"fmt"
	"hash"
	"os"
)

func hashByAlgo(algo string) (hash.Hash, error) {
	switch algo {
	case "md5":
		return md5.New(), nil
	case "sha256":
		return sha256.New(), nil
	case "sha512":
		return sha512.New(), nil
	}
	return nil, fmt.Errorf("unknown algorithm %q (use md5|sha256|sha512)", algo)
}

func computeHex(algo, data string) (string, error) {
	h, err := hashByAlgo(algo)
	if err != nil {
		return "", err
	}
	_, _ = h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil)), nil
}

func equalFoldHex(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		ca, cb := a[i], b[i]
		if 'A' <= ca && ca <= 'Z' {
			ca += 'a' - 'A'
		}
		if 'A' <= cb && cb <= 'Z' {
			cb += 'a' - 'A'
		}
		if ca != cb {
			return false
		}
	}
	return true
}

func usage() {
	fmt.Fprint(os.Stderr, `hash — утилита хэширования (MD5 | SHA-256 | SHA-512)

Использование:
  hash -algo=sha256 -text "hello"
  hash -algo=sha256 -text "hello" -check=<hex>

Флаги:
  -algo    md5 | sha256 | sha512 (по умолчанию sha256)
  -text    строка для обработки
  -check   если указан — сравнить хэш строки с указанным hex
`)
}

func main() {
	algo := flag.String("algo", "sha256", "md5|sha256|sha512")
	text := flag.String("text", "", "строка для обработки")
	check := flag.String("check", "", "hex-хэш для проверки целостности")
	flag.Usage = usage
	flag.Parse()

	if *text == "" {
		usage()
		os.Exit(2)
	}

	sum, err := computeHex(*algo, *text)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	if *check != "" {
		fmt.Printf("algo:      %s\n", *algo)
		fmt.Printf("computed:  %s\n", sum)
		fmt.Printf("provided:  %s\n", *check)
		if equalFoldHex(sum, *check) {
			fmt.Println("result:    OK — хэши совпадают")
			return
		}
		fmt.Println("result:    MISMATCH — данные повреждены или хэш неверный")
		os.Exit(3)
	}
	fmt.Println(sum)
}
