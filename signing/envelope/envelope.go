package envelope

import (
	"encoding/json"

	"github.com/secure-systems-lab/go-securesystemslib/dsse"
)

/*
Envelope captures an envelope as described by the Secure Systems Lab
Signing Specification. See here:
https://github.com/secure-systems-lab/signing-spec/blob/master/envelope.md
*/
type Envelope struct {
	PayloadType string      `json:"payloadType"`
	Payload     string      `json:"payload"`
	Signatures  []Signature `json:"signatures"`
}

/*
Signature represents a generic in-toto signature that contains the identifier
of the key which was used to create the signature.
The used signature scheme has to be agreed upon by the signer and verifer
out of band.
The signature is a base64 encoding of the raw bytes from the signature
algorithm.
The cert is a PEM encoded string of the signing certificate
*/
type Signature struct {
	KeyID string `json:"keyid"`
	Sig   string `json:"sig"`
	Cert  string `json:"cert"`
}

func AddCertToEnvelope(signedAtt []byte, cert []byte) ([]byte, error) {
	// Unmarshal into a DSSE envelope.
	env := &dsse.Envelope{}
	if err := json.Unmarshal(signedAtt, env); err != nil {
		return nil, err
	}

	// Create an envelope.Envelope.
	envWithCert := &Envelope{
		PayloadType: env.PayloadType,
		Payload:     env.Payload,
		Signatures:  []Signature{},
	}

	for _, sig := range env.Signatures {
		envWithCert.Signatures = append(envWithCert.Signatures,
			Signature{Sig: sig.Sig, KeyID: sig.KeyID, Cert: string(cert)})
	}

	// Return marshalled result
	return json.Marshal(envWithCert)
}
