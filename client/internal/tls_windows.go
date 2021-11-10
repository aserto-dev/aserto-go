package internal

import (
	"crypto/tls"
)

func TLSConfig(insecure bool) (*tls.Config, error) {
	var tlsConf tls.Config

	if insecure {
		tlsConf.InsecureSkipVerify = true
	} else {
		tlsConf.MinVersion = tls.VersionTLS12
	}
	return &tlsConf, nil
}
