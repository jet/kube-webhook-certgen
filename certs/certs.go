package certs

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/pem"
	"github.com/pkg/errors"
	"math/big"
	"net"
	"time"
)

// GenerateCerts creates ed25519 ca with a leaf certificate. The returned values are the ca cert and leaf cert, leaf
// private key. The ca key is discarded.
func GenerateCerts(hosts []string) (ca []byte, cert []byte, key []byte, err error) {
	notBefore := time.Now().Add(time.Minute * -5)
	notAfter := notBefore.Add(100 * 365 * 24 * time.Hour)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "failed to generate serial number")
	}
	rootPub, rootPriv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "failed ed25519.GenerateKey")
	}

	rootTemplate := x509.Certificate{
		SerialNumber:          serialNumber,
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &rootTemplate, &rootTemplate, rootPub, rootPriv)
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "failed createCertificate for Ca")
	}

	ca, err = encodeCert(derBytes)
	if err != nil {
		return nil, nil, nil, err
	}

	leafPub, leafPriv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "failed ed25519.GenerateKey")
	}

	key, err = encodeKey(leafPriv)
	if err != nil {
		return nil, nil, nil, err
	}

	serialNumber, err = rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "failed to generate serial number")
	}
	leafTemplate := x509.Certificate{
		SerialNumber:          serialNumber,
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  false,
	}
	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			leafTemplate.IPAddresses = append(leafTemplate.IPAddresses, ip)
		} else {
			leafTemplate.DNSNames = append(leafTemplate.DNSNames, h)
		}
	}

	derBytes, err = x509.CreateCertificate(rand.Reader, &leafTemplate, &rootTemplate, leafPub, rootPriv)
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "failed createLeaf certificate")
	}

	cert, err = encodeCert(derBytes)
	return ca, cert, key, err
}

func encodeKey(key ed25519.PrivateKey) ([]byte, error) {
	// Marshal the innter CurvePrivateKey.
	derBytes, err := encodeEd25519(key)
	if err != nil {
		return nil, err
	}
	encoded := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: derBytes})
	if encoded == nil {
		return nil, errors.New("unable to encode key to pem")
	}
	return encoded, nil
}

func encodeCert(derBytes []byte) ([]byte, error) {
	encoded := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	if encoded == nil {
		return nil, errors.New("unable to encode cert to pem")
	}
	return encoded, nil
}

// https://github.com/cloudflare/cfssl/blob/master/helpers/derhelpers/ed25519.go
func encodeEd25519(key ed25519.PrivateKey) ([]byte, error) {
	curvePrivateKey, err := asn1.Marshal(key.Seed())
	if err != nil {
		return nil, err
	}

	return asn1.Marshal(
		struct {
			Version    int
			Algorithm  pkix.AlgorithmIdentifier
			PrivateKey []byte
		}{
			0,
			pkix.AlgorithmIdentifier{Algorithm: asn1.ObjectIdentifier{1, 3, 101, 112}},
			curvePrivateKey,
		})
}
