package main

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"flag"
	"fmt"
	"os"
)

func generateKeys(privPath, pubPath string) error {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}
	privDER, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return err
	}
	privPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privDER})
	if err := os.WriteFile(privPath, privPEM, 0o600); err != nil {
		return err
	}

	pubDER, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	if err != nil {
		return err
	}
	pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubDER})
	return os.WriteFile(pubPath, pubPEM, 0o644)
}

func loadPrivateKey(path string) (*rsa.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("no PEM block in %s", path)
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA private key")
	}
	return rsaKey, nil
}

func loadPublicKey(path string) (*rsa.PublicKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("no PEM block in %s", path)
	}
	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	rsaPub, ok := key.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA public key")
	}
	return rsaPub, nil
}

func sign(priv *rsa.PrivateKey, message []byte) ([]byte, error) {
	digest := sha256.Sum256(message)
	return rsa.SignPKCS1v15(rand.Reader, priv, crypto.SHA256, digest[:])
}

func verify(pub *rsa.PublicKey, message, signature []byte) error {
	digest := sha256.Sum256(message)
	return rsa.VerifyPKCS1v15(pub, crypto.SHA256, digest[:], signature)
}

func usage() {
	fmt.Fprint(os.Stderr, `rsa — цифровая подпись (RSA-2048 + SHA-256)

  rsa -mode=gen    -priv=priv.pem -pub=pub.pem
  rsa -mode=sign   -priv=priv.pem -text "hello"
  rsa -mode=verify -pub=pub.pem  -text "hello" -sig=<base64>
`)
}

func main() {
	mode := flag.String("mode", "gen", "gen | sign | verify")
	privPath := flag.String("priv", "priv.pem", "файл приватного ключа")
	pubPath := flag.String("pub", "pub.pem", "файл публичного ключа")
	text := flag.String("text", "", "сообщение")
	sigB64 := flag.String("sig", "", "base64-подпись")
	flag.Usage = usage
	flag.Parse()

	switch *mode {
	case "gen":
		if err := generateKeys(*privPath, *pubPath); err != nil {
			fmt.Fprintln(os.Stderr, "gen:", err)
			os.Exit(1)
		}
		fmt.Printf("OK: %s, %s\n", *privPath, *pubPath)

	case "sign":
		priv, err := loadPrivateKey(*privPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, "load priv:", err)
			os.Exit(1)
		}
		sig, err := sign(priv, []byte(*text))
		if err != nil {
			fmt.Fprintln(os.Stderr, "sign:", err)
			os.Exit(1)
		}
		fmt.Println(base64.StdEncoding.EncodeToString(sig))

	case "verify":
		pub, err := loadPublicKey(*pubPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, "load pub:", err)
			os.Exit(1)
		}
		sig, err := base64.StdEncoding.DecodeString(*sigB64)
		if err != nil {
			fmt.Fprintln(os.Stderr, "bad sig:", err)
			os.Exit(1)
		}
		if err := verify(pub, []byte(*text), sig); err != nil {
			fmt.Println("подпись НЕВЕРНА:", err)
			os.Exit(3)
		}
		fmt.Println("подпись верна")

	default:
		usage()
		os.Exit(2)
	}
}
