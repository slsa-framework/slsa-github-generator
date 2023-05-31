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

const process = require("process");
import * as types from "../src/predicate";

const predicate = require("../src/predicate");

describe("generatePredicate", () => {
  const originalEnv = process.env;

  beforeEach(() => {
    jest.resetModules();
    process.env = {
      GITHUB_WORKFLOW_REF: "asraa/slsa-github-generator/.github/workflows/pre-submit.e2e.container-based.default.yml@refs/heads/main"
    };
  });

  afterEach(() => {
    process.env = originalEnv;
  });

  it("get predicate with resolveddependencies", () => {
    const jobWorkflowRef =
      "octo-org/octo-automation/.github/workflows/oidc.yml@refs/heads/main";

    const builderBinaryRef = {
      uri: "git+https://github.com/slsa-framework/example-builder",
      digest: {
        sha256: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
      },
    };

    const currentRun = {
      id: 123456,
      run_attempt: 1,
      repository: {
        name: "example-source",
        full_name: "slsa-framework/example-source",
        id: 123,
        owner: {
          id: 456
        }
      },
      actor: {
        id: 456
      },
      triggering_actor: {
        id: 456
      }
    }

    const source = {
      "uri": "git+https://github.com/slsa-framework/example-source",
      "digest": {
        "sha1": "97f1bfd54b02d1c7b632da907676a7d30d2efc02"
      }
    }

    const bd = {
      buildType: "https://slsa.dev/container-based-build/v0.1?draft",
      externalParameters: {
        "source": source
      },
      resolvedDependencies: [source]
    }


    const pred = predicate.generatePredicate(bd, builderBinaryRef, jobWorkflowRef, currentRun) as types.SLSAv1Predicate;

    expect(pred.runDetails.builder.id).toBe(jobWorkflowRef);
    expect(pred.buildDefinition.resolvedDependencies).toHaveLength(2);
  });
});
