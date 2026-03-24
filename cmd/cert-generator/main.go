package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"log"
	"math/big"
	"net"
	"os"
	"time"
)

func main() {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatal(err)
	}

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		log.Fatal(err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Goshortener Dev"},
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(365 * 24 * time.Hour),

		KeyUsage: x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth,
		},

		IPAddresses: []net.IP{net.ParseIP("127.0.0.1")},
		DNSNames:    []string{"localhost"},
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		log.Fatal(err)
	}

	certOut, err := os.Create("cert.pem")
	if err != nil {
		log.Fatal(err)
	}
	defer certOut.Close()

	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})

	keyOut, err := os.Create("key.pem")
	if err != nil {
		log.Fatal(err)
	}
	defer keyOut.Close()

	pem.Encode(keyOut, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(priv),
	})

	log.Println("cert.pem and key.pem generated")
}
