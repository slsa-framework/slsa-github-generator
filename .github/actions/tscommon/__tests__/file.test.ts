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

const file = require("../src/file");
const path = require("path");
const fs = require("fs");
const wd = file.getGitHubWorkspace();

beforeAll(() => {
  if (!file.safeExistsSync("safeunlink")) {
    file.safeWriteFileSync("safeunlink", "data");
  }
  if (!file.safeExistsSync("safermdir")) {
    file.safeMkdirSync("safermdir");
  }
  initGitHub();
});

afterAll(() => {
  const tmpFiles: string[] = ["safefilesha256", "safewritefile"];
  for (const fn of tmpFiles) {
    if (file.safeExistsSync(fn)) {
      file.safeUnlinkSync(fn);
    }
  }

  const tmpDirs: string[] = ["safemkdir"];
  for (const dn of tmpDirs) {
    if (file.safeExistsSync(dn)) {
      file.safeRmdirSync(dn);
    }
  }

  cleanupGitHub();
});

function initGitHub(): void {
  const isCI = process.env.CI || "";
  if (isCI) {
    return;
  }
  // NOTE: for local testing.
  const eventPath = process.env.GITHUB_EVENT_PATH || "";
  const eventDir = path.dirname(eventPath);
  if (!fs.existsSync(eventDir)) {
    fs.mkdirSync(eventDir, { recursive: true });
  }
  if (!fs.existsSync(eventPath)) {
    fs.writeFileSync(eventPath, "data");
  }
  const runnerTmp = process.env.RUNNER_TEMP || "";
  if (!fs.existsSync(runnerTmp)) {
    fs.mkdirSync(runnerTmp, { recursive: true });
  }
}

function cleanupGitHub(): void {
  const isCI = process.env.CI || "";
  if (isCI) {
    return;
  }
  // NOTE: for local testing.
  const eventPath = process.env.GITHUB_EVENT_PATH || "";
  if (fs.existsSync(eventPath)) {
    fs.unlinkSync(eventPath);
  }
  const runnerTmp = process.env.RUNNER_TEMP || "";
  if (fs.existsSync(runnerTmp)) {
    fs.rmdirSync(runnerTmp, { recursive: true, force: true });
  }
}

describe("safeFileSha256", () => {
  const tmpFiles: string[] = [
    "safefilesha256",
    path.join(process.env.RUNNER_TEMP || "", "file"),
    "/tmp/safewritefile",
  ];

  beforeEach(() => {
    for (const fn of tmpFiles) {
      fs.writeFileSync(fn, "some data", {
        flag: "wx",
        mode: 0o600,
      });
    }
  });

  afterEach(() => {
    for (const fn of tmpFiles) {
      if (fs.existsSync(fn)) {
        fs.unlinkSync(fn);
      }
    }
  });

  it("calculates the file sha", () => {
    const input = "safefilesha256";
    const expected =
      "1307990e6ba5ca145eb35e99182a9bec46531bc54ddf656a602c780fa0240dee";

    expect(file.safeFileSha256(input)).toEqual(expected);
  });

  it("fails on path traversal", () => {
    const input = "../path";
    expect(() => file.safeFileSha256(input)).toThrow(Error);
  });

  it("fails on path traversal with same prefix", () => {
    const input = wd + "path";
    expect(() => file.safeFileSha256(input)).toThrow(Error);
  });

  it("fails on path traversal with trailing slash", () => {
    const input = "../path/";
    expect(() => file.safeFileSha256(input)).toThrow(Error);
  });

  it("fails on path traversal with join", () => {
    const input = path.join(wd, "../path");
    expect(() => file.safeFileSha256(input)).toThrow(Error);
  });

  it("fails on path traversal of /tmp", () => {
    const input = "/tmp/../safewritefile";
    expect(() => file.safeFileSha256(input, "data")).toThrow(Error);
  });

  it("fails on path traversal of RUNNER_TEMP", () => {
    const input = path.join(process.env.RUNNER_TEMP || "", "..", "file");
    expect(() => file.safeFileSha256(input, "data")).toThrow(Error);
  });

  it("calculates sha for file in RUNNER_TEMP", () => {
    const input = path.join(process.env.RUNNER_TEMP || "", "file");
    const expected =
      "1307990e6ba5ca145eb35e99182a9bec46531bc54ddf656a602c780fa0240dee";
    expect(file.safeFileSha256(input)).toEqual(expected);
  });

  it("calculates sha for file in /tmp", () => {
    const input = "/tmp/safewritefile";
    const expected =
      "1307990e6ba5ca145eb35e99182a9bec46531bc54ddf656a602c780fa0240dee";
    expect(file.safeFileSha256(input)).toEqual(expected);
  });

  it("calculates the sha of event path data", () => {
    const input = process.env.GITHUB_EVENT_PATH || "";
    expect(file.safeFileSha256(input)).toBeTruthy();
  });

  it("fails on path traversal tmp with same prefix", () => {
    const input = "/tmppath";
    expect(() => file.safeFileSha256(input)).toThrow(Error);
  });

  it("fails on path traversal with RUNNER_TEMP same prefix", () => {
    const input = (process.env.RUNNER_TEMP || "") + "path";
    expect(() => file.safeFileSha256(input)).toThrow(Error);
  });
});

describe("resolvePathInput", () => {
  beforeEach(() => {
    if (fs.existsSync("/tmp/hello")) {
      fs.unlinkSync("/tmp/hello");
    }
  });

  it("path traversal", () => {
    const input = "../path";
    expect(() => file.resolvePathInput(input, true)).toThrow(Error);
    expect(() => file.resolvePathInput(input, false)).toThrow(Error);
  });

  it("path traversal same start", () => {
    const input = wd + "path";
    expect(() => file.resolvePathInput(input, true)).toThrow(Error);
    expect(() => file.resolvePathInput(input, false)).toThrow(Error);
  });

  it("path traversal with trailing", () => {
    const input = "../path/";
    expect(() => file.resolvePathInput(input, true)).toThrow(Error);
    expect(() => file.resolvePathInput(input, false)).toThrow(Error);
  });

  it("path traversal with join", () => {
    const input = path.join(wd, "../path-other");
    expect(() => file.resolvePathInput(input, true)).toThrow(Error);
    expect(() => file.resolvePathInput(input, false)).toThrow(Error);
  });

  it("safe path traversal twice", () => {
    const input = "path";
    const safe = file.resolvePathInput(input, true);
    expect(safe).toEqual(`${wd}/path`);
    const safesafe = file.resolvePathInput(safe, true);
    expect(safesafe).toEqual(`${safe}`);

    const input2 = "path2";
    const safe2 = file.resolvePathInput(input2, false);
    expect(safe2).toEqual(`${wd}/path2`);
    const safesafe2 = file.resolvePathInput(safe2, false);
    expect(safesafe2).toEqual(`${safe2}`);
  });

  it("safe path /tmp/bla", () => {
    const input = "/tmp/bla";
    expect(file.resolvePathInput(input, true)).toEqual(input);

    const input2 = "/tmp/bla2";
    expect(file.resolvePathInput(input2, false)).toEqual(input2);
  });

  it("safe event file", () => {
    const input = process.env.GITHUB_EVENT_PATH || "";
    expect(file.resolvePathInput(input, false)).toEqual(input);
    expect(() => file.resolvePathInput(input, true)).toThrow(Error);
  });

  it("path traversal with tmp", () => {
    const input = "/tmp/../hello";
    expect(() => file.resolvePathInput(input, true)).toThrow(Error);
    expect(() => file.resolvePathInput(input, false)).toThrow(Error);
  });

  it("safe with tmp", () => {
    const input = "/tmp/hello";
    expect(file.resolvePathInput(input, true)).toEqual(input);
    expect(file.resolvePathInput(input, false)).toEqual(input);
  });

  it("path traversal with runner tmp", () => {
    const input = path.join(process.env.RUNNER_TEMP || "", "..", "file");
    console.log(`input: ${input}`);
    expect(() => file.resolvePathInput(input, true)).toThrow(Error);
    expect(() => file.resolvePathInput(input, false)).toThrow(Error);
  });

  it("safe with runner tmp", () => {
    const input = path.join(process.env.RUNNER_TEMP || "", "file");
    expect(file.resolvePathInput(input, true)).toEqual(input);
    expect(file.resolvePathInput(input, false)).toEqual(input);
  });

  it("path traversal tmp same start", () => {
    const input = "/tmppath";
    expect(() => file.resolvePathInput(input, true)).toThrow(Error);
    expect(() => file.resolvePathInput(input, false)).toThrow(Error);
  });

  it("path traversal runner tmp same start", () => {
    const input = (process.env.RUNNER_TEMP || "") + "path";
    expect(() => file.resolvePathInput(input, true)).toThrow(Error);
    expect(() => file.resolvePathInput(input, false)).toThrow(Error);
  });
});

describe("safeWriteFileSync", () => {
  beforeEach(() => {
    const fn = path.join(process.env.RUNNER_TEMP || "", "file");
    if (fs.existsSync(fn)) {
      fs.unlinkSync(fn);
    }
  });

  it("path traversal", () => {
    const input = "../path";
    expect(() => file.safeWriteFileSync(input, "data")).toThrow(Error);
  });

  it("path traversal same start", () => {
    const input = wd + "path";
    expect(() => file.safeWriteFileSync(input, "data")).toThrow(Error);
  });

  it("path traversal with trailing", () => {
    const input = "../path/";
    expect(() => file.safeWriteFileSync(input, "data")).toThrow(Error);
  });

  it("safe path", () => {
    const input = "safewritefile";
    file.safeWriteFileSync(input, "data");
  });

  it("path traversal with join", () => {
    const input = path.join(wd, "..", "path");
    expect(() => file.safeWriteFileSync(input, "data")).toThrow(Error);
  });

  it("path traversal overwrite", () => {
    const input = "safewritefile";
    expect(() => file.safeWriteFileSync(input, "data")).toThrow();
  });

  it("path traversal /tmp", () => {
    const input = "/tmp/../safewritefile";
    expect(() => file.safeWriteFileSync(input, "data")).toThrow();
  });

  it("path traversal runner tmp", () => {
    const input = path.join(process.env.RUNNER_TEMP || "", "..", "file");
    expect(() => file.safeWriteFileSync(input, "data")).toThrow();
  });

  it("event path write", () => {
    const input = process.env.GITHUB_EVENT_PATH || "";
    expect(() => file.safeWriteFileSync(input, "data")).toThrow();
  });

  it("safe with tmp", () => {
    const input = "/tmp/hello";
    file.safeWriteFileSync(input, "data");
    expect(() => file.safeWriteFileSync(input, "data")).toThrow();
  });

  it("safe with runner tmp", () => {
    const input = path.join(process.env.RUNNER_TEMP || "", "file");
    file.safeWriteFileSync(input, "data");
    expect(() => file.safeWriteFileSync(input, "data")).toThrow();
  });

  it("path traversal tmp same start", () => {
    const input = "/tmppath";
    expect(() => file.safeWriteFileSync(input)).toThrow(Error);
  });

  it("path traversal runner tmp same start", () => {
    const input = (process.env.RUNNER_TEMP || "") + "path";
    expect(() => file.safeWriteFileSync(input)).toThrow(Error);
  });
});

describe("safeReadFileSync", () => {
  beforeEach(() => {
    const files: string[] = [
      path.join(process.env.RUNNER_TEMP || "", "file"),
      "/tmp/safewritefile",
    ];
    for (const fn of files) {
      if (fs.existsSync(fn)) {
        fs.unlinkSync(fn);
      }
    }
  });

  it("path traversal", () => {
    const input = "../path";
    expect(() => file.safeReadFileSync(input)).toThrow(Error);
  });

  it("path traversal same start", () => {
    const input = wd + "path";
    expect(() => file.safeReadFileSync(input)).toThrow(Error);
  });

  it("path traversal with trailing", () => {
    const input = "../path/";
    expect(() => file.safeReadFileSync(input)).toThrow(Error);
  });

  it("path traversal with join", () => {
    const input = path.join(wd, "../path");
    expect(() => file.safeReadFileSync(input)).toThrow(Error);
  });

  it("safe path", () => {
    const input = "README.md";
    const content = file.safeReadFileSync(input);
  });

  it("path traversal /tmp", () => {
    const input = "/tmp/../safewritefile";
    expect(() => file.safeReadFileSync(input, "data")).toThrow();
  });

  it("safe tmp", () => {
    const input = "/tmp/safewritefile";
    file.safeWriteFileSync(input, "data");
    expect(file.safeReadFileSync(input).toString()).toEqual("data");
  });

  it("path traversal runner tmp", () => {
    const input = path.join(process.env.RUNNER_TEMP || "", "..", "file");
    expect(() => file.safeReadFileSync(input, "data")).toThrow();
  });

  it("safe runner tmp", () => {
    const input = path.join(process.env.RUNNER_TEMP || "", "file");
    file.safeWriteFileSync(input, "data");
    expect(file.safeReadFileSync(input).toString()).toEqual("data");
  });

  it("event path", () => {
    const input = process.env.GITHUB_EVENT_PATH || "";
    file.safeReadFileSync(input);
  });

  it("path traversal tmp same start", () => {
    const input = "/tmppath";
    expect(() => file.safeReadFileSync(input)).toThrow(Error);
  });

  it("path traversal runner tmp same start", () => {
    const input = (process.env.RUNNER_TEMP || "") + "path";
    expect(() => file.safeReadFileSync(input)).toThrow(Error);
  });
});

describe("safeMkdirSync", () => {
  beforeEach(() => {
    const dirs: string[] = [
      "/tmp/safedir",
      path.join(process.env.RUNNER_TEMP || "", "dir"),
    ];
    for (const d of dirs) {
      if (fs.existsSync(d)) {
        fs.rmdirSync(d);
      }
    }
  });

  it("path traversal", () => {
    const input = "../path";
    expect(() => file.safeMkdirSync(input)).toThrow(Error);
  });

  it("path traversal same start", () => {
    const input = wd + "path";
    expect(() => file.safeMkdirSync(input)).toThrow(Error);
  });

  it("path traversal with trailing", () => {
    const input = "../path/";
    expect(() => file.safeMkdirSync(input)).toThrow(Error);
  });

  it("path traversal with join", () => {
    const input = path.join(wd, "../path");
    expect(() => file.safeMkdirSync(input)).toThrow(Error);
  });

  it("safe path traversal", () => {
    const input = "safemkdir";
    file.safeMkdirSync(input);
  });

  it("safe path traversal overwrite", () => {
    const input = "safemkdir";
    expect(() => file.safeMkdirSync(input)).toThrow();
  });

  it("path traversal /tmp", () => {
    const input = "/tmp/../dir";
    expect(() => file.safeMkdirSync(input)).toThrow();
  });

  it("safe tmp", () => {
    const input = "/tmp/safedir";
    file.safeMkdirSync(input);
  });

  it("path traversal runner tmp", () => {
    const input = path.join(process.env.RUNNER_TEMP || "", "..", "file");
    expect(() => file.safeMkdirSync(input)).toThrow();
  });

  it("safe runner tmp", () => {
    const input = path.join(process.env.RUNNER_TEMP || "", "dir");
    file.safeMkdirSync(input);
  });

  it("event path", () => {
    const input = process.env.GITHUB_EVENT_PATH || "";
    expect(() => file.safeMkdirSync(input)).toThrow();
  });

  it("path traversal tmp same start", () => {
    const input = "/tmppath";
    expect(() => file.safeMkdirSync(input)).toThrow(Error);
  });

  it("path traversal runner tmp same start", () => {
    const input = (process.env.RUNNER_TEMP || "") + "path";
    expect(() => file.safeMkdirSync(input)).toThrow(Error);
  });
});

describe("safeUnlinkSync", () => {
  it("path traversal", () => {
    const input = "../path";
    expect(() => file.safeUnlinkSync(input)).toThrow(Error);
  });

  it("path traversal same start", () => {
    const input = wd + "path";
    expect(() => file.safeUnlinkSync(input)).toThrow(Error);
  });

  it("path traversal with trailing", () => {
    const input = "../path/";
    expect(() => file.safeUnlinkSync(input)).toThrow(Error);
  });

  it("path traversal with join", () => {
    const input = path.join(wd, "../path-other");
    expect(() => file.safeUnlinkSync(input)).toThrow(Error);
  });

  it("safe path", () => {
    const input = "safeunlink";
    file.safeUnlinkSync(input);
  });

  it("path traversal /tmp", () => {
    const input = "/tmp/../file";
    expect(() => file.safeUnlinkSync(input)).toThrow();
  });

  it("safe tmp", () => {
    const input = "/tmp/safefile";
    fs.writeFileSync(input, "data");
    file.safeUnlinkSync(input);
  });

  it("path traversal runner tmp", () => {
    const input = path.join(process.env.RUNNER_TEMP || "", "..", "file");
    expect(() => file.safeUnlinkSync(input)).toThrow();
  });

  it("safe runner tmp", () => {
    const input = path.join(process.env.RUNNER_TEMP || "", "file");
    fs.writeFileSync(input, "data");
    file.safeUnlinkSync(input);
  });

  it("event path", () => {
    const input = process.env.GITHUB_EVENT_PATH || "";
    expect(() => file.safeUnlinkSync(input)).toThrow();
  });

  it("path traversal tmp same start", () => {
    const input = "/tmppath";
    expect(() => file.safeUnlinkSync(input)).toThrow(Error);
  });

  it("path traversal runner tmp same start", () => {
    const input = (process.env.RUNNER_TEMP || "") + "path";
    expect(() => file.safeUnlinkSync(input)).toThrow(Error);
  });
});

describe("safeRmdirSync", () => {
  it("path traversal", () => {
    const input = "../path";
    expect(() => file.safeRmdirSync(input)).toThrow(Error);
  });

  it("path traversal same start", () => {
    const input = wd + "path";
    expect(() => file.safeRmdirSync(input)).toThrow(Error);
  });

  it("path traversal with trailing", () => {
    const input = "../path/";
    expect(() => file.safeRmdirSync(input)).toThrow(Error);
  });

  it("path traversal with join", () => {
    const input = path.join(wd, "../path-other");
    expect(() => file.safeRmdirSync(input)).toThrow(Error);
  });

  it("safe path", () => {
    const input = "safermdir";
    file.safeRmdirSync(input);
  });

  it("safe path traversal not present", () => {
    const input = "safermdir";
    expect(() => file.safeRmdirSync(input)).toThrow();
  });

  it("path traversal /tmp", () => {
    const input = "/tmp/../dir";
    expect(() => file.safeRmdirSync(input)).toThrow();
  });

  it("safe tmp", () => {
    const input = "/tmp/adir";
    fs.mkdirSync("/tmp/adir");
    file.safeRmdirSync(input);
  });

  it("path traversal runner tmp", () => {
    const input = path.join(process.env.RUNNER_TEMP || "", "..", "file");
    expect(() => file.safeRmdirSync(input)).toThrow();
  });

  it("safe runner tmp", () => {
    const input = path.join(process.env.RUNNER_TEMP || "", "bdir");
    fs.mkdirSync(input);
    file.safeRmdirSync(input);
  });

  it("event path", () => {
    const input = process.env.GITHUB_EVENT_PATH || "";
    expect(() => file.safeRmdirSync(input)).toThrow();
  });

  it("path traversal tmp same start", () => {
    const input = "/tmppath";
    expect(() => file.safeRmdirSync(input)).toThrow(Error);
  });

  it("path traversal runner tmp same start", () => {
    const input = (process.env.RUNNER_TEMP || "") + "path";
    expect(() => file.safeRmdirSync(input)).toThrow(Error);
  });
});

describe("safeExistsSync", () => {
  it("path traversal", () => {
    const input = "../path";
    expect(() => file.safeExistsSync(input)).toThrow(Error);
  });

  it("path traversal same start", () => {
    const input = wd + "path";
    expect(() => file.safeExistsSync(input)).toThrow(Error);
  });

  it("path traversal with trailing", () => {
    const input = "../path/";
    expect(() => file.safeExistsSync(input)).toThrow(Error);
  });

  it("path traversal with join", () => {
    const input = path.join(wd, "../path-other");
    expect(() => file.safeExistsSync(input)).toThrow(Error);
  });

  it("safe path traversal", () => {
    const input = "README.md";
    file.safeExistsSync(input);
  });

  it("safe path traversal not present", () => {
    const input = "README.md.not.here";
    expect(() => file.safeRmdirSync(input)).toThrow();
  });

  it("path traversal /tmp", () => {
    const input = "/tmp/../file";
    expect(() => file.safeExistsSync(input)).toThrow();
  });

  it("safe tmp", () => {
    const input = "/tmp/safefile";
    file.safeExistsSync(input);
  });

  it("path traversal runner tmp", () => {
    const input = path.join(process.env.RUNNER_TEMP || "", "../file");
    expect(() => file.safeExistsSync(input)).toThrow();
  });

  it("safe runner tmp", () => {
    const input = path.join(process.env.RUNNER_TEMP || "", "file");
    file.safeExistsSync(input);
  });

  it("event path", () => {
    const input = process.env.GITHUB_EVENT_PATH || "";
    file.safeExistsSync(input);
  });

  it("path traversal tmp same start", () => {
    const input = "/tmppath";
    expect(() => file.safeExistsSync(input)).toThrow(Error);
  });

  it("path traversal runner tmp same start", () => {
    const input = (process.env.RUNNER_TEMP || "") + "path";
    expect(() => file.safeExistsSync(input)).toThrow(Error);
  });
});

describe("safePromises_readdir", () => {
  beforeEach(() => {
    const dirs: string[] = [
      "/tmp/readdir",
      path.join(process.env.RUNNER_TEMP || "", "readdir"),
    ];
    for (const d of dirs) {
      if (!fs.existsSync(d)) {
        fs.mkdirSync(d);
      }
    }
  });

  it("path traversal", async () => {
    const input = "../path";
    await expect(file.safePromises_readdir(input)).rejects.toThrow();
  });

  it("path traversal same start", async () => {
    const input = wd + "path";
    await expect(file.safePromises_readdir(input)).rejects.toThrow();
  });

  it("path traversal with trailing", async () => {
    const input = "../path/";
    await expect(file.safePromises_readdir(input)).rejects.toThrow();
  });

  it("path traversal with join", async () => {
    const input = path.join(wd, "../path-other");
    await expect(file.safePromises_readdir(input)).rejects.toThrow();
  });

  it("safe path traversal", async () => {
    const input = "src/";
    file.safePromises_readdir(input);
  });

  it("safe path traversal not present", async () => {
    const input = "not-present";
    await expect(file.safePromises_readdir(input)).rejects.toThrow();
  });

  it("path traversal /tmp", async () => {
    const input = "/tmp/../file";
    await expect(file.safePromises_readdir(input)).rejects.toThrow();
  });

  it("safe tmp", () => {
    const input = "/tmp/readdir";
    file.safePromises_readdir(input);
  });

  it("path traversal runner tmp", async () => {
    const input = path.join(process.env.RUNNER_TEMP || "", "..", "file");
    await expect(file.safePromises_readdir(input)).rejects.toThrow();
  });

  it("safe runner tmp", () => {
    const input = path.join(process.env.RUNNER_TEMP || "", "readdir");
    file.safePromises_readdir(input);
  });

  it("event path", async () => {
    const input = process.env.GITHUB_EVENT_PATH || "";
    // NOTE: not a directory.
    await expect(file.safePromises_readdir(input)).rejects.toThrow();
  });

  it("path traversal tmp same start", async () => {
    const input = "/tmppath";
    await expect(file.safePromises_readdir(input)).rejects.toThrow();
  });

  it("path traversal runner tmp same start", async () => {
    const input = (process.env.RUNNER_TEMP || "") + "path";
    await expect(file.safePromises_readdir(input)).rejects.toThrow();
  });
});

describe("safePromises_stat", () => {
  beforeEach(() => {
    const files: string[] = [
      "/tmp/safefile",
      path.join(process.env.RUNNER_TEMP || "", "file"),
    ];
    for (const f of files) {
      if (!fs.existsSync(f)) {
        fs.writeFileSync(f, "data");
      }
    }
  });

  it("path traversal", async () => {
    const input = "../path";
    await expect(file.safePromises_stat(input)).rejects.toThrow();
  });

  it("path traversal same start", async () => {
    const input = wd + "path";
    await expect(file.safePromises_stat(input)).rejects.toThrow();
  });

  it("path traversal with trailing", async () => {
    const input = "../path/";
    await expect(file.safePromises_stat(input)).rejects.toThrow();
  });

  it("path traversal with join", async () => {
    const input = path.join(wd, "../path-other");
    await expect(file.safePromises_stat(input)).rejects.toThrow();
  });

  it("safe path traversal", async () => {
    const input = "src/";
    file.safePromises_stat(input);
  });

  it("safe path traversal not present", async () => {
    const input = "not-present";
    await expect(file.safePromises_stat(input)).rejects.toThrow();
  });

  it("path traversal /tmp", async () => {
    const input = "/tmp/../file";
    await expect(file.safePromises_stat(input)).rejects.toThrow();
  });

  it("safe tmp", () => {
    const input = "/tmp/safefile";
    file.safePromises_stat(input);
  });

  it("path traversal runner tmp", async () => {
    const input = path.join(process.env.RUNNER_TEMP || "", "..", "file");
    await expect(file.safePromises_stat(input)).rejects.toThrow();
  });

  it("safe runner tmp", () => {
    const input = path.join(process.env.RUNNER_TEMP || "", "file");
    file.safePromises_stat(input);
  });

  it("event path", async () => {
    const input = process.env.GITHUB_EVENT_PATH || "";
    // NOTE: not a directory.
    await expect(file.safePromises_stat(input)).rejects.toThrow();
  });

  it("path traversal tmp same start", async () => {
    const input = "/tmppath";
    await expect(file.safePromises_stat(input)).rejects.toThrow();
  });

  it("path traversal runner tmp same start", async () => {
    const input = (process.env.RUNNER_TEMP || "") + "path";
    await expect(file.safePromises_stat(input)).rejects.toThrow();
  });

  it("path traversal runner tmp same start", async () => {
    const input = (process.env.RUNNER_TEMP || "") + "path";
    await expect(file.safePromises_stat(input)).rejects.toThrow();
  });
});
