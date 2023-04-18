package sigstore

import (
	"context"
	"fmt"

	"github.com/sigstore/cosign/v2/pkg/cosign"
	"github.com/sigstore/rekor/pkg/client"
	"github.com/sigstore/rekor/pkg/generated/client/entries"
	"github.com/sigstore/rekor/pkg/generated/models"
	"github.com/slsa-framework/slsa-github-generator/signing"
)

const (
	// DefaultRekorAddr is the default rekor base URL.
	DefaultRekorAddr = "https://rekor.sigstore.dev"
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

// NewRekor returns a new Rekor instance.
func NewRekor(rekorAddr string) *Rekor {
	return &Rekor{
		rekorAddr: rekorAddr,
	}
}

// Upload uploads the signed attestation to the rekor transparency log.
func (r *Rekor) Upload(ctx context.Context, att signing.Attestation) (signing.LogEntry, error) {
	rekorClient, err := client.GetRekorClient(r.rekorAddr)
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
		entry := entry
		pubs, err := cosign.GetRekorPubs(ctx)
		if err != nil {
			return nil, fmt.Errorf("getting rekor public keys: %w", err)
		}

		if err := cosign.VerifyTLogEntryOffline(ctx, &entry, pubs); err != nil {
			return nil, fmt.Errorf("validating log entry: %w", err)
		}
		uuid = ix
		logEntry = &entry
	}

	fmt.Printf("Uploaded signed attestation to rekor with UUID %s.\n", uuid)
	fmt.Printf("You could use rekor-cli to view the log entry details:\n\n"+
		"  $ rekor-cli get --uuid %[1]s\n\n"+
		"In addition to that, you could also use the Rekor Search UI:\n\n"+
		"  https://search.sigstore.dev/?uuid=%[1]s", uuid)
	return &rekorEntryAnon{
		entry: logEntry,
		uuid:  uuid,
	}, nil
}
