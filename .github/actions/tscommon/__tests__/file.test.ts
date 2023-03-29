const file = require("../src/file");
const path = require("path");
const wd = file.getGitHubWorkspace()

beforeAll(() => {
    if (file.safeExistsSync("safewritefile")){
        file.safeUnlinkSync("safewritefile")
    }
    if (file.safeExistsSync("safemkdir")){
        file.rmdirSync("safemkdir")
    }
    if (!file.safeExistsSync("safeunlink")){
        file.safeWriteFileSync("safeunlink", "data")
    }
    if (!file.safeExistsSync("safermdir")){
        file.safeMkdirSync("safermdir")
    }
});

describe("resolvePathInput", () => {
  it("path traversal", () => {
    const input = "../path";
    expect(() => file.resolvePathInput(input)).toThrow(Error);
  });

  it("path traversal with trailing", () => {
    const input = "../path/";
    expect(() => file.resolvePathInput(input)).toThrow(Error);
  });

  it("path traversal with join", () => {
    const input = path.join(wd, "../path-other");
    expect(() => file.resolvePathInput(input)).toThrow(Error);
  });

  it("safe path traversal", () => {
    const input = "path";
    const safe = file.resolvePathInput(input);
    expect(safe).toEqual(`${wd}/path`);
    const safesafe = file.resolvePathInput(safe);
    expect(safesafe).toEqual(`${safe}`);
  });
});

describe("safeWriteFileSync", () => {
    it("path traversal", () => {
        const input = "../path";
        expect(() => file.safeWriteFileSync(input, "data")).toThrow(Error);
    });

    it("path traversal with trailing", () => {
        const input = "../path/";
        expect(() => file.safeWriteFileSync(input, "data")).toThrow(Error);
    });

    it("safe path traversal", () => {
        const input = "safewritefile";
        file.safeWriteFileSync(input, "data");
    });

    it("path traversal with join", () => {
        const input = path.join(wd, "../path");
        expect(() => file.safeWriteFileSync(input, "data")).toThrow(Error);
    });

    it("safe path traversal overwrite", () => {
        const input = "safewritefile";
        expect(() => file.safeWriteFileSync(input, "data")).toThrow();
    });
});

describe("safeReadFileSync", () => {
    it("path traversal", () => {
        const input = "../path";
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

    it("safe path traversal", () => {
        const input = "README.md";
        file.safeReadFileSync(input);
    });
});

describe("safeMkdirSync", () => {
    it("path traversal", () => {
        const input = "../path";
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
});

describe("safeUnlinkSync", () => {
    it("path traversal", () => {
        const input = "../path";
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

    it("safe path traversal", () => {
        const input = "safeunlink";
        file.safeUnlinkSync(input);
    });

    it("safe path traversal not present", () => {
        const input = "safemkdir";
        expect(() => file.safeMkdirSync(input)).toThrow();
    });
});

describe("rmdirSync", () => {
    it("path traversal", () => {
        const input = "../path";
        expect(() => file.rmdirSync(input)).toThrow(Error);
    });

    it("path traversal with trailing", () => {
        const input = "../path/";
        expect(() => file.rmdirSync(input)).toThrow(Error);
    });

    it("path traversal with join", () => {
        const input = path.join(wd, "../path-other");
        expect(() => file.rmdirSync(input)).toThrow(Error);
    });

    it("safe path traversal", () => {
        const input = "safermdir";
        file.rmdirSync(input);
    });

    it("safe path traversal not present", () => {
        const input = "safermdir";
        expect(() => file.rmdirSync(input)).toThrow();
    });
});

describe("safeExistsSync", () => {
    it("path traversal", () => {
        const input = "../path";
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
        expect(() => file.rmdirSync(input)).toThrow();
    });
});

describe("safePromises_readdir", () => {
    it("path traversal", async () => {
        const input = "../path";
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
});

describe("safePromises_stat", () => {
    it("path traversal", async () => {
        const input = "../path";
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
});