// // Copyright 2023 SLSA Authors
// //
// // Licensed under the Apache License, Version 2.0 (the "License");
// // you may not use this file except in compliance with the License.
// // You may obtain a copy of the License at
// //
// //      http://www.apache.org/licenses/LICENSE-2.0
// //
// // Unless required by applicable law or agreed to in writing, software
// // distributed under the License is distributed on an "AS IS" BASIS,
// // WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// // See the License for the specific language governing permissions and
// // limitations under the License.

const labels = require("../src/labels");
const main = require("../src/main");
const core = require("@actions/core");

jest.mock("@actions/core");
jest.mock("@octokit/rest", () => ({ Octokit: jest.fn() }));
jest.mock("../src/labels");

beforeEach(() => {
  jest.resetModules();
  labels.getJobLabels.mockClear();
  labels.getSelfHostedRunnerLabels.mockClear();
  core.getInput.mockClear();
  core.setFailed.mockClear();
  core.info.mockClear();
});

describe("getUsedSelfHostedRunnerLabels", () => {
  const [token, owner, repo] = ["any", "any", "any"];

  it("failure: can't get jobs", async () => {
    const msg = "can't get jobs";
    // console.dir(detect)
    labels.getJobLabels.mockRejectedValue(new Error(msg));
    labels.getSelfHostedRunnerLabels.mockReturnValue(new Set([]));
    expect(
      main.getUsedSelfHostedRunnerLabels(token, owner, repo),
    ).rejects.toThrow(new Error(msg));
  });

  it("failure: can't get self-hosted runners", async () => {
    const msg = "can't get runners";
    // console.dir(detect)
    labels.getJobLabels.mockReturnValue(new Set([]));
    labels.getSelfHostedRunnerLabels.mockRejectedValue(new Error(msg));
    expect(
      main.getUsedSelfHostedRunnerLabels(token, owner, repo),
    ).rejects.toThrow(new Error(msg));
  });

  it("success: no common labels", () => {
    labels.getJobLabels.mockReturnValue(new Set(["abc"]));
    labels.getSelfHostedRunnerLabels.mockReturnValue(new Set(["def"]));
    expect(
      main.getUsedSelfHostedRunnerLabels(token, owner, repo),
    ).resolves.toEqual([]);
  });

  it("success: some common labels", () => {
    labels.getJobLabels.mockReturnValue(new Set(["abc", "xyz"]));
    labels.getSelfHostedRunnerLabels.mockReturnValue(new Set(["def", "xyz"]));
    expect(
      main.getUsedSelfHostedRunnerLabels(token, owner, repo),
    ).resolves.toEqual(["xyz"]);
  });
});

describe("run", () => {
  process.env.GITHUB_REPOSITORY = "any/any";

  it("failure: can't get labels", async () => {
    const msg = "any error";
    labels.getJobLabels.mockRejectedValue(new Error(msg));
    await main.run();
    expect(core.getInput).toHaveBeenCalledWith("token");
    expect(core.setFailed).toHaveBeenCalledWith(`error: ${msg}`);
    // expect(core.info).not.toHaveBeenCalled();
  });

  it("failure: self-hosted runners are used", async () => {
    labels.getJobLabels.mockReturnValue(new Set(["abc", "xyz"]));
    labels.getSelfHostedRunnerLabels.mockReturnValue(
      new Set(["abc", "def", "xyz"]),
    );
    const msg = `Self-hosted runners are not allowed in SLSA Level 3 workflows. labels: abc,xyz`;
    await main.run();
    expect(core.getInput).toHaveBeenCalledWith("token");
    expect(core.setFailed).toHaveBeenCalledWith(msg);
    expect(core.info).not.toHaveBeenCalled();
  });

  it("success: no use of self-hosted runners", async () => {
    labels.getJobLabels.mockReturnValue(new Set(["abc"]));
    labels.getSelfHostedRunnerLabels.mockReturnValue(new Set(["def"]));
    const msg = `Self-hosted runners are not allowed in SLSA Level 3 workflows. labels: abc,xyz`;
    await main.run();
    expect(core.getInput).toHaveBeenCalledWith("token");
    expect(core.setFailed).not.toHaveBeenCalled();
    expect(core.info).toHaveBeenCalledWith("No self-hosted runners detected");
  });
});
