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

const core = require("@actions/core");
const detect = require("../src/detect");
const main = require("../src/main");
const repo = "slsa-framework/gundam";

jest.mock("../src/detect");
jest.mock("@actions/core");

beforeEach(() => {
  jest.resetModules();
  process.env.GITHUB_REPOSITORY = repo;
  core.setFailed.mockClear();
  core.setOutput.mockClear();
  detect.ensureOnlyGithubHostedRunners.mockClear();
  detect.detectWorkflowFromContext.mockClear();
});

describe("main.run", () =>{
  const [repo, ref, builderWorkflow] = ["abc", "123", "xyz"];
  const genericGeneratorWorkflow = ".github/workflows/generator_generic_slsa3.yml";

  it("run succeeds: workflow is NOT generic generator", async () => {
    detect.detectWorkflowFromContext.mockReturnValue(Promise.resolve([repo, ref, builderWorkflow]));

    await main.run();

    expect(detect.detectWorkflowFromContext).toHaveBeenCalled();
    expect(detect.ensureOnlyGithubHostedRunners).not.toHaveBeenCalled();

    expect(core.setFailed).not.toHaveBeenCalled();

    expect(core.setOutput).toHaveBeenCalledWith("repository", repo);
    expect(core.setOutput).toHaveBeenCalledWith("ref", ref);
    expect(core.setOutput).toHaveBeenCalledWith("workflow", builderWorkflow);
  });


  it("run succeeds: workflow is a generic generator, no self-hosted runners", async () => {
    detect.detectWorkflowFromContext.mockReturnValue(Promise.resolve([repo, ref, genericGeneratorWorkflow]));
    detect.ensureOnlyGithubHostedRunners.mockReturnValue(Promise.resolve(null));

    await main.run();

    expect(detect.detectWorkflowFromContext).toHaveBeenCalled();
    expect(detect.ensureOnlyGithubHostedRunners).toHaveBeenCalled();

    expect(core.setFailed).not.toHaveBeenCalled();

    expect(core.setOutput).toHaveBeenCalledWith("repository", repo);
    expect(core.setOutput).toHaveBeenCalledWith("ref", ref);
    expect(core.setOutput).toHaveBeenCalledWith("workflow", genericGeneratorWorkflow);
  });

  it("run fails: can't get workflow details", async () => {
    const errMsg = "can't get workflow details";
    detect.detectWorkflowFromContext.mockRejectedValue(new Error(errMsg));
    detect.ensureOnlyGithubHostedRunners.mockReturnValue(Promise.resolve(null));

    await main.run();

    expect(detect.detectWorkflowFromContext).toHaveBeenCalled();
    expect(detect.ensureOnlyGithubHostedRunners).not.toHaveBeenCalled();

    expect(core.setFailed).toHaveBeenCalledWith(errMsg);

    expect(core.setOutput).not.toHaveBeenCalled();
    expect(core.setOutput).not.toHaveBeenCalled();
    expect(core.setOutput).not.toHaveBeenCalled();
  });

  it("run fails: generic workflow, but using self-hosted runner", async () => {
    const errMsg = "no self-hosted runners allowed";
    detect.detectWorkflowFromContext.mockReturnValue(Promise.resolve([repo, ref, genericGeneratorWorkflow]));
    detect.ensureOnlyGithubHostedRunners.mockRejectedValue(new Error(errMsg));

    await main.run();

    expect(detect.detectWorkflowFromContext).toHaveBeenCalled();
    expect(detect.ensureOnlyGithubHostedRunners).toHaveBeenCalled();

    expect(core.setFailed).toHaveBeenCalledWith(errMsg);

    expect(core.setOutput).toHaveBeenCalledWith("repository", repo);
    expect(core.setOutput).toHaveBeenCalledWith("ref", ref);
    expect(core.setOutput).toHaveBeenCalledWith("workflow", genericGeneratorWorkflow);
  });
});
