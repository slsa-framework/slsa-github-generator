package envelope

import (
	"bytes"
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"math/big"
	"testing"

	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
	"github.com/in-toto/in-toto-golang/in_toto"
	"github.com/secure-systems-lab/go-securesystemslib/dsse"
	sdsse "github.com/sigstore/sigstore/pkg/signature/dsse"

	intoto "github.com/in-toto/in-toto-golang/in_toto"
	"github.com/sigstore/rekor/pkg/generated/models"
	"github.com/sigstore/rekor/pkg/types"
	intotod "github.com/sigstore/rekor/pkg/types/intoto/v0.0.1"
	"github.com/sigstore/sigstore/pkg/signature"
)

func intotoEntry(certPem, provenance []byte) (*intotod.V001Entry, error) {
	cert := strfmt.Base64(certPem)
	return &intotod.V001Entry{
		IntotoObj: models.IntotoV001Schema{
			Content: &models.IntotoV001SchemaContent{
				Envelope: string(provenance),
			},
			PublicKey: &cert,
		},
	}, nil
}

// marshals a dsse envelope for testing.
func marshalEnvelope(t *testing.T, env *dsse.Envelope) string {
	b, err := json.Marshal(env)
	if err != nil {
		t.Fatalf("marshalling envelope: %s", err)
	}
	return string(b)
}

// test utility to sign a payload with a given signer.
func envelope(t *testing.T, k *ecdsa.PrivateKey, payload []byte) string {
	s, err := signature.LoadECDSASigner(k, crypto.SHA256)
	if err != nil {
		t.Fatal(err)
	}
	wrappedSigner := sdsse.WrapSigner(s, intoto.PayloadType)
	if err != nil {
		t.Fatal(err)
	}
	dsseEnv, err := wrappedSigner.SignMessage(bytes.NewReader(payload))
	if err != nil {
		t.Fatal(err)
	}
	return string(dsseEnv)
}

func TestAddCert(t *testing.T) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	ca := &x509.Certificate{
		SerialNumber: big.NewInt(1),
	}
	caBytes, err := x509.CreateCertificate(rand.Reader, ca, ca, &priv.PublicKey, priv)
	if err != nil {
		t.Fatal(err)
	}
	certPemBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	})
	validPayload := "hellothispayloadisvalid"

	tests := []struct {
		name   string
		env    string
		cert   []byte
		addErr bool
	}{
		{
			name:   "invalid empty envelope with no signatures",
			env:    marshalEnvelope(t, &dsse.Envelope{}),
			cert:   nil,
			addErr: true,
		},
		{
			name: "invalid envelope with two signatures",
			env: marshalEnvelope(t, &dsse.Envelope{
				Payload:     "",
				PayloadType: in_toto.PayloadType,
				Signatures: []dsse.Signature{
					{
						Sig: "abc",
					},
					{
						Sig: "xyz",
					},
				},
			}),
			cert:   nil,
			addErr: true,
		},
		{
			name: "invalid cert with valid envelope",
			env: marshalEnvelope(t, &dsse.Envelope{
				Payload:     "",
				PayloadType: in_toto.PayloadType,
				Signatures: []dsse.Signature{
					{
						Sig: "abc",
					},
				},
			}),
			cert:   nil,
			addErr: true,
		},
		{
			name:   "valid envelope",
			env:    envelope(t, priv, []byte(validPayload)),
			cert:   certPemBytes,
			addErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Add certificate to envelope.
			envWithCert, err := AddCertToEnvelope([]byte(tt.env), tt.cert)
			if (err != nil) != tt.addErr {
				t.Errorf("AddCertToEnvelope() error = %v, wanted %v", err, tt.addErr)
			}
			if err != nil {
				return
			}

			// Now get cert from envelope and compare.
			gotCert, err := GetCertFromEnvelope(envWithCert)
			if err != nil {
				t.Fatalf("GetCertFromEnvelope() error = %v", err)
			}

			if !bytes.EqualFold(gotCert, tt.cert) {
				t.Errorf("expected cert equality")
			}

			// Now test compatibility with Rekor intoto entry type.
			testRekorSupport(t, tt.cert, envWithCert)
		})
	}
}

// This servers as a regression test to make sure that the Rekor intoto
// type can successfully unmarshal our "Envelope" with included cert.
func testRekorSupport(t *testing.T, certPem, envWithCert []byte) {
	ctx := context.Background()
	intotoEntry, err := intotoEntry(certPem, envWithCert)
	if err != nil {
		t.Fatalf("error creating intoto entry: %s", err)
	}
	e := models.Intoto{
		APIVersion: swag.String(intotoEntry.APIVersion()),
		Spec:       intotoEntry.IntotoObj,
	}
	pe := models.ProposedEntry(&e)
	entry, err := types.CreateVersionedEntry(pe)
	if err != nil {
		t.Fatalf("error creating valid intoto entry")
	}
	_, err = types.CanonicalizeEntry(ctx, entry)
	if err != nil {
		t.Fatalf("error creating valid intoto entry")
	}
}
