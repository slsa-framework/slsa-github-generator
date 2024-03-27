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

const labels = require("../src/labels");
const Octokit = require("@octokit/rest");

function createOctokitMock() {
  return {
    rest: {
      actions: {
        listJobsForWorkflowRun: jest.fn(),
        listSelfHostedRunnersForRepo: jest.fn(),
      },
    },
    paginate: jest.fn(),
  };
}

jest.mock("@octokit/rest", () => ({ Octokit: jest.fn() }));

beforeEach(() => {
  jest.resetModules();
  Octokit.Octokit.mockClear();
});

describe("getJobLabels", () => {
  const octokit = createOctokitMock();
  Octokit.Octokit.mockReturnValue(octokit);

  const owner = "anyowner";
  const repo = "anyrepo";
  const run_id = 123;
  process.env.GITHUB_RUN_ID = run_id.toString();

  it("failure: api error", async () => {
    const msg = "any";
    octokit.paginate.mockRejectedValue(new Error(msg));

    expect(labels.getJobLabels(octokit, owner, repo)).rejects.toThrow(
      new Error(msg),
    );
    expect(octokit.paginate).toHaveBeenCalledWith(
      octokit.rest.actions.listJobsForWorkflowRun,
      {
        owner,
        repo,
        run_id,
      },
    );
  });

  it("success", () => {
    octokit.paginate.mockReturnValue(
      Promise.resolve([
        {
          id: 399444496,
          run_id: 29679449,
          name: "myjob",
          labels: ["foo", "bar"],
        },
        {
          id: 399444497,
          run_id: 29679449,
          name: "myotherjob",
          labels: ["baz"],
        },
      ]),
    );
    expect(labels.getJobLabels(octokit, owner, repo)).resolves.toEqual(
      new Set(["foo", "bar", "baz"]),
    );
  });
});

describe("getSelfHostedRunnerLabels", () => {
  const octokit = createOctokitMock();
  Octokit.Octokit.mockReturnValue(octokit);

  const owner = "anyowner";
  const repo = "anyrepo";

  it("failure: api error", async () => {
    const msg = "any";
    octokit.paginate.mockRejectedValue(new Error(msg));

    expect(
      labels.getSelfHostedRunnerLabels(octokit, owner, repo),
    ).rejects.toThrow(new Error(msg));
    expect(octokit.paginate).toHaveBeenCalledWith(
      octokit.rest.actions.listSelfHostedRunnersForRepo,
      {
        owner,
        repo,
      },
    );
  });

  it("success: zero self-hosted runners", () => {
    octokit.paginate.mockReturnValue(Promise.resolve([]));
    expect(
      labels.getSelfHostedRunnerLabels(octokit, owner, repo),
    ).resolves.toEqual(new Set([]));
  });

  it("success: some self-hosted runners", () => {
    octokit.paginate.mockReturnValue(
      Promise.resolve([
        {
          id: 24,
          name: "my-gh-runner",
          os: "Linux",
          status: "online",
          busy: true,
          labels: [
            { id: 1, name: "self-hosted", type: "read-only" },
            { id: 2, name: "Linux", type: "read-only" },
            { id: 3, name: "X64", type: "read-only" },
            { id: 7, name: "gundam", type: "custom" },
          ],
        },
        {
          id: 25,
          name: "my-laptop",
          os: "Linux",
          status: "offline",
          busy: false,
          labels: [
            { id: 1, name: "self-hosted", type: "read-only" },
            { id: 2, name: "Linux", type: "read-only" },
            { id: 3, name: "X64", type: "read-only" },
          ],
        },
      ]),
    );
    expect(
      labels.getSelfHostedRunnerLabels(octokit, owner, repo),
    ).resolves.toEqual(new Set(["self-hosted", "Linux", "X64", "gundam"]));
  });
});
