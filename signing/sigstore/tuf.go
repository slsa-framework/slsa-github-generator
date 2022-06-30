package sigstore

import _ "embed"

// go:embed staging-root.json
var StagingRoot []byte

const (
	StagingTufAddr = "https://storage.googleapis.com/tuf-root-staging"
)
