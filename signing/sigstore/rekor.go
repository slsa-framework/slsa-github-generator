package sigstore

import (
	"context"
	"fmt"

	"github.com/sigstore/cosign/cmd/cosign/cli/rekor"
	"github.com/sigstore/cosign/pkg/cosign"
	"github.com/sigstore/rekor/pkg/generated/models"
	"github.com/slsa-framework/slsa-github-generator/signing"
)

const (
	defaultRekorAddr = "https://rekor.sigstore.dev"
)

// Rekor implements TransparencyLog
type Rekor struct {
	rekorAddr string
}

type rekorEntryAnon struct {
	entry *models.LogEntryAnon
}

// ID implements LogEntry.ID.
func (e *rekorEntryAnon) ID() string {
	if e.entry.LogID == nil {
		return ""
	}
	return *e.entry.LogID
}

// NewDefaultRekor returns a new Rekor instance for the Rekor public instance.
func NewDefaultRekor() *Rekor {
	return NewRekor(defaultRekorAddr)
}

// NewRekor returns a new Rekor instance.
func NewRekor(rekorAddr string) *Rekor {
	return &Rekor{
		rekorAddr: rekorAddr,
	}
}

// Upload uploads the signed attestation to the rekor transparency log.
func (r *Rekor) Upload(ctx context.Context, att signing.Attestation) (signing.LogEntry, error) {
	rekorClient, err := rekor.NewClient(r.rekorAddr)
	if err != nil {
		return nil, fmt.Errorf("creating rekor client: %w", err)
	}
	// TODO: Is it a bug that we need []byte(string(k.Cert)) or else we hit invalid PEM?
	logEntry, err := cosign.TLogUploadInTotoAttestation(ctx, rekorClient, att.Bytes(), []byte(string(att.Cert())))
	if err != nil {
		return nil, fmt.Errorf("uploading attestation: %w", err)
	}

	return &rekorEntryAnon{
		entry: logEntry,
	}, nil
}
