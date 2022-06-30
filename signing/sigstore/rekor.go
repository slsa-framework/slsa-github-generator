package sigstore

import (
	"context"
	"fmt"

	"github.com/sigstore/cosign/cmd/cosign/cli/rekor"
	"github.com/sigstore/cosign/pkg/cosign"
	"github.com/sigstore/rekor/pkg/generated/client/entries"
	"github.com/sigstore/rekor/pkg/generated/models"
	"github.com/slsa-framework/slsa-github-generator/signing"
)

const (
	DefaultRekorAddr = "https://rekor.sigstore.dev"
	StagingRekorAddr = "https://rekor.sigstage.dev"
)

// Rekor implements TransparencyLog.
type Rekor struct {
	rekorAddr string
}

type rekorEntryAnon struct {
	entry *models.LogEntryAnon
	uuid  string
}

// ID implements LogEntry.ID.
func (e *rekorEntryAnon) ID() string {
	if e.entry.LogID == nil {
		return ""
	}
	return *e.entry.LogID
}

// LogIndex implements LogEntry.LogIndex.
func (e *rekorEntryAnon) LogIndex() int64 {
	if e.entry.LogIndex == nil {
		return -1
	}
	return *e.entry.LogIndex
}

// UUID implements LogEntry.UUID.
func (e *rekorEntryAnon) UUID() string {
	return e.uuid
}

// NewDefaultRekor returns a new Rekor instance for the Rekor public instance.
func NewDefaultRekor() *Rekor {
	return NewRekor(DefaultRekorAddr)
}

// NewStagingRekor returns a new Rekor instance for the Rekor staging instance.
func NewStagingRekor() *Rekor {
	return NewRekor(StagingRekorAddr)
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

	params := entries.NewGetLogEntryByIndexParamsWithContext(ctx)
	params.SetLogIndex(*logEntry.LogIndex)
	resp, err := rekorClient.Entries.GetLogEntryByIndex(params)
	if err != nil {
		return nil, fmt.Errorf("retrieving log uuid by index: %w", err)
	}
	var uuid string
	for ix, entry := range resp.Payload {
		if err := cosign.VerifyTLogEntry(ctx, rekorClient, &entry); err != nil {
			return nil, fmt.Errorf("validating log entry: %w", err)
		}
		uuid = ix
		logEntry = &entry
	}

	fmt.Printf("Uploaded signed attestation to rekor with UUID %s.\n", uuid)
	return &rekorEntryAnon{
		entry: logEntry,
		uuid:  uuid,
	}, nil
}
