package common

import (
	"context"
	"encoding/json"
	"os"

	"github.com/slsa-framework/slsa-github-generator/internal/utils"
	"github.com/slsa-framework/slsa-github-generator/slsa"
)

// Generate generates a SLSA predicate and writes it to the given path.
func Generate(provider slsa.ClientProvider, bt *GenericBuild, path string) error {
	ctx := context.Background()

	if provider != nil {
		bt.WithClients(provider)
	} else {
		// TODO(github.com/slsa-framework/slsa-github-generator/issues/124): Remove
		if utils.IsPresubmitTests() {
			bt.WithClients(&slsa.NilClientProvider{})
		}
	}

	g := slsa.NewHostedActionsGenerator(bt)
	if provider != nil {
		g.WithClients(provider)
	} else {
		// TODO(github.com/slsa-framework/slsa-github-generator/issues/124): Remove
		if utils.IsPresubmitTests() {
			g.WithClients(&slsa.NilClientProvider{})
		}
	}

	p, err := g.Generate(ctx)
	if err != nil {
		return err
	}

	pb, err := json.Marshal(p.Predicate)
	if err != nil {
		return err
	}

	pf, err := utils.CreateNewFileUnderCurrentDirectory(path, os.O_WRONLY)
	if err != nil {
		return err
	}

	_, err = pf.Write(pb)
	return err
}
