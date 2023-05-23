// Copyright 2023 SLSA Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package signing

import (
	"context"

	intoto "github.com/in-toto/in-toto-golang/in_toto"
)

// Attestation is a signed attestation.
type Attestation interface {
	// Cert returns the certificate used to sign the attestation.
	Cert() []byte

	// Bytes returns the signed attestation as an encoded DSSE JSON envelope.
	Bytes() []byte
}

// Signer is used to sign provenance statements.
type Signer interface {
	// Sign signs the given provenance statement and returns the signed
	// attestation.
	Sign(context.Context, *intoto.Statement) (Attestation, error)
}

// LogEntry represents a transparency log entry.
type LogEntry interface {
	// ID returns the ID of the transparency log.
	ID() string

	// LogIndex return the index of the transparency log entry.
	LogIndex() int64

	// UUID return the uuid of the transparency log entry.
	UUID() string
}

// TransparencyLog allows interaction with a transparency log.
type TransparencyLog interface {
	// Upload uploads the signed attestation to the transparency log.
	Upload(context.Context, Attestation) (LogEntry, error)
}
