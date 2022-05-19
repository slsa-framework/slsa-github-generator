// Copyright 2022 SLSA Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
