package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"github.com/slsa-framework/slsa-github-generator/github"
)

func main() {
	ctx := context.Background()

	c, err := github.NewOIDCClient()
	if err != nil {
		log.Fatal(err)
	}

	audience := os.Getenv("GITHUB_REPOSITORY")
	if audience == "" {
		log.Fatal("missing github repository environment context")
	}
	audience = path.Join(audience, "detect-workflow")

	t, err := c.Token(ctx, []string{audience})
	if err != nil {
		log.Fatal(err)
	}

	pathParts := strings.SplitN(t.JobWorkflowRef, "/", 3)
	if len(pathParts) < 3 {
		log.Fatal("missing org/repository in job workflow ref")
	}
	repository := strings.Join(pathParts[:2], "/")

	refParts := strings.Split(t.JobWorkflowRef, "@")
	if len(refParts) < 2 {
		log.Fatal("missing reference in job workflow ref")
	}

	fmt.Println(fmt.Sprintf(`::set-output name=repository::%s`, repository))
	fmt.Println(fmt.Sprintf(`::set-output name=ref::%s`, refParts[1]))
}
