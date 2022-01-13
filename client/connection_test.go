package client // nolint:testpackage

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/pem"
	"fmt"
	"math/big"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

type dialRecorder struct {
	context     context.Context
	address     string
	tlsConf     *tls.Config
	callerCreds credentials.PerRPCCredentials
	connection  *Connection
}

func (d *dialRecorder) DialContext(
	ctx context.Context,
	address string,
	tlsConf *tls.Config,
	callerCreds credentials.PerRPCCredentials,
	connection *Connection,
) (grpc.ClientConnInterface, error) {
	d.context = ctx
	d.address = address
	d.tlsConf = tlsConf
	d.callerCreds = callerCreds
	d.connection = connection

	return &grpc.ClientConn{}, nil
}

func TestWithAddr(t *testing.T) {
	recorder := &dialRecorder{}
	newConnection(context.TODO(), recorder.DialContext, WithAddr("address")) // nolint:errcheck

	assert.Equal(t, "address", recorder.address)
}

func TestWithURL(t *testing.T) {
	recorder := &dialRecorder{}

	const URL = "https://server.com:123"
	svcURL, err := url.Parse(URL)
	assert.NoError(t, err)

	newConnection(context.TODO(), recorder.DialContext, WithURL(svcURL)) // nolint:errcheck

	assert.Equal(t, URL, recorder.address)
}

func TestAddrAndURL(t *testing.T) {
	recorder := &dialRecorder{}

	svcURL, err := url.Parse("https://server.com:123")
	assert.NoError(t, err)

	_, err = newConnection(context.TODO(), recorder.DialContext, WithAddr("address"), WithURL(svcURL))
	assert.Error(t, err)
}

func TestWithInsecure(t *testing.T) {
	recorder := &dialRecorder{}
	newConnection(context.TODO(), recorder.DialContext, WithInsecure(true)) // nolint:errcheck

	assert.True(t, recorder.tlsConf.InsecureSkipVerify)
}

func TestWithTokenAuth(t *testing.T) {
	recorder := &dialRecorder{}
	newConnection(context.TODO(), recorder.DialContext, WithTokenAuth("<token>")) // nolint:errcheck

	md, err := recorder.callerCreds.GetRequestMetadata(context.TODO())
	assert.NoError(t, err)

	token, ok := md["authorization"]
	assert.True(t, ok)
	assert.Equal(t, "bearer <token>", token)
}

func TestWithAPIKey(t *testing.T) {
	recorder := &dialRecorder{}
	newConnection(context.TODO(), recorder.DialContext, WithAPIKeyAuth("<apikey>")) // nolint:errcheck

	md, err := recorder.callerCreds.GetRequestMetadata(context.TODO())
	assert.NoError(t, err)

	token, ok := md["authorization"]
	assert.True(t, ok)
	assert.Equal(t, "basic <apikey>", token)
}

func TestTokenAndAPIKey(t *testing.T) {
	recorder := &dialRecorder{}

	_, err := newConnection(context.TODO(), recorder.DialContext, WithAPIKeyAuth("<apikey>"), WithTokenAuth("<token>"))
	assert.Error(t, err)
}

func TestWithTenantID(t *testing.T) {
	recorder := &dialRecorder{}
	newConnection(context.TODO(), recorder.DialContext, WithTenantID("<tenantid>")) // nolint:errcheck

	assert.Equal(t, "<tenantid>", recorder.connection.TenantID)

	ctx := context.TODO()
	recorder.connection.unary( // nolint:errcheck
		ctx,
		"method",
		"request",
		"reply",
		recorder.connection.Conn.(*grpc.ClientConn),
		func(c context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
			md, ok := metadata.FromOutgoingContext(c)
			assert.True(t, ok)

			tenantID := md.Get("aserto-tenant-id")
			assert.Equal(t, 1, len(tenantID), "request should contain tenant ID metadata field")
			assert.Equal(t, "<tenantid>", tenantID[0], "tenant ID metadata should have the expected value")

			assert.Equal(t, "method", method, "'method' parameter should be a passthrough")
			assert.Equal(t, "request", req, "'request' parameter should be a passthrough")
			assert.Equal(t, "reply", reply, "'reply' parameter should be a passthrough")
			assert.Equal(t, recorder.connection.Conn, cc)

			return nil
		})

	recorder.connection.stream( // nolint:errcheck
		ctx,
		nil,
		recorder.connection.Conn.(*grpc.ClientConn),
		"method",
		func(
			c context.Context,
			desc *grpc.StreamDesc,
			cc *grpc.ClientConn,
			method string,
			opts ...grpc.CallOption,
		) (grpc.ClientStream, error) {
			md, ok := metadata.FromOutgoingContext(c)
			assert.True(t, ok)

			tenantID := md.Get("aserto-tenant-id")
			assert.Equal(t, 1, len(tenantID), "request should contain tenant ID metadata field")
			assert.Equal(t, "<tenantid>", tenantID[0], "tenant ID metadata should have the expected value")

			assert.Equal(t, "method", method, "'method' parameter should be a passthrough")
			assert.Equal(t, recorder.connection.Conn, cc)

			return nil, nil
		},
	)
}

const CertSubjectName = "Testing Inc."

func TestWithCACertPath(t *testing.T) {
	tempdir := t.TempDir()
	caPath := fmt.Sprintf("%s/ca.pem", tempdir)

	file, err := os.OpenFile(caPath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)
	assert.NoError(t, err, "Failed to create CA file")

	defer file.Close()

	caCert, err := generateCACert(CertSubjectName)
	assert.NoError(t, err, "Failed to generate test certificate")

	_, err = file.Write(caCert)
	assert.NoError(t, err, "Failed to save certificate")

	recorder := &dialRecorder{}
	newConnection(context.TODO(), recorder.DialContext, WithCACertPath(caPath)) // nolint:errcheck

	inPool, err := subjectInCertPool(recorder.tlsConf.RootCAs, CertSubjectName)
	if err != nil {
		t.Errorf("Failed to read root CAs: %s", err)
	}

	assert.True(t, inPool, "Aserto cert should be in root CAs")
}

func TestWithCACertPathAndInsecure(t *testing.T) {
	tempdir := t.TempDir()
	caPath := fmt.Sprintf("%s/ca.pem", tempdir)

	file, err := os.OpenFile(caPath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)
	assert.NoError(t, err, "Failed to create CA file")

	defer file.Close()

	caCert, err := generateCACert(CertSubjectName)
	assert.NoError(t, err, "Failed to generate test certificate")

	_, err = file.Write(caCert)
	assert.NoError(t, err, "Failed to save certificate")

	recorder := &dialRecorder{}
	newConnection(context.TODO(), recorder.DialContext, WithCACertPath(caPath), WithInsecure(true)) // nolint:errcheck

	assert.Nil(t, recorder.tlsConf.RootCAs, "Aserto cert should be nil")
	assert.True(t, recorder.tlsConf.InsecureSkipVerify)
}

func generateCACert(subjectName string) ([]byte, error) {
	priv, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	if err != nil {
		return nil, err
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{subjectName},
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(time.Hour * 24 * 180),

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, publicKey(priv), priv)
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate: %w", err)
	}

	out := &bytes.Buffer{}
	if err := pem.Encode(out, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return nil, fmt.Errorf("Failed to PEM encode certificate: %w", err)
	}

	return out.Bytes(), nil
}

func publicKey(priv interface{}) interface{} {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &k.PublicKey
	case *ecdsa.PrivateKey:
		return &k.PublicKey
	default:
		return nil
	}
}

func subjectInCertPool(pool *x509.CertPool, name string) (bool, error) {
	for _, subject := range pool.Subjects() {
		var rdns pkix.RDNSequence

		_, err := asn1.Unmarshal(subject, &rdns)
		if err != nil {
			return false, err
		}

		for _, nameset := range rdns {
			for _, attr := range nameset {
				if attr.Value == name {
					return true, nil
				}
			}
		}
	}

	return false, nil
}
