package tlsconf

import (
	"crypto/tls"
	"crypto/x509"
	"os"

	"github.com/pkg/errors"
)

func TLSConfig(insecure bool, caCertPath string) (*tls.Config, error) {
	var (
		tlsConf  tls.Config
		certPool *x509.CertPool
		err      error
	)

	if insecure {
		tlsConf.InsecureSkipVerify = true
		return &tlsConf, nil
	}

	if caCertPath == "" {
		certPool, err = x509.SystemCertPool()
		if err != nil {
			return nil, errors.Wrap(err, "failed to get system cert pool")
		}
	} else {
		certPool = x509.NewCertPool()
		caCertBytes, err := os.ReadFile(caCertPath)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to read ca cert [%s]", caCertPath)
		}

		if !certPool.AppendCertsFromPEM(caCertBytes) {
			return nil, errors.Wrapf(err, "failed to append client ca cert [%s]", caCertPath)
		}
	}

	tlsConf.RootCAs = certPool
	tlsConf.MinVersion = tls.VersionTLS12

	return &tlsConf, nil
}
