package internal

import (
	"crypto/tls"
	"crypto/x509"

	"github.com/pkg/errors"
)

func TLSConfig(insecure bool) (*tls.Config, error) {
	var tlsConf tls.Config

	if insecure {
		tlsConf.InsecureSkipVerify = true
	} else {
		certPool, err := x509.SystemCertPool()
		if err != nil {
			return nil, errors.Wrap(err, "failed to get system cert pool")
		}
		tlsConf.RootCAs = certPool
		tlsConf.MinVersion = tls.VersionTLS12
	}

	return &tlsConf, nil
}
