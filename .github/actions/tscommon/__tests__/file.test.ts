const file = require("../src/file");
const wd = process.env["GITHUB_WORKSPACE"] || process.env["PWD"] || "";

describe("resolvePathInput", () => {
  it("path traversal", () => {
    const input = "../path";
    expect(() => file.resolvePathInput(input)).toThrow(Error);
  });

  it("path traversal with trailing", () => {
    const input = "../path/";
    expect(() => file.resolvePathInput(input)).toThrow(Error);
  });

  it("path traversal with same start", () => {
    const input = "../path-other";
    expect(() => file.resolvePathInput(input)).toThrow(Error);
  });

  it("safe path traversal", () => {
    const input = "path";
    const safe = file.resolvePathInput(input);
    expect(safe).toEqual(`${wd}/path`);
  });
});

jest.mock("fs");

describe("writeFileSync", () => {
    it("path traversal", () => {
        const input = "../path";
        expect(() => file.writeFileSync(input, "data")).toThrow(Error);
    });

    it("path traversal with trailing", () => {
        const input = "../path/";
        expect(() => file.writeFileSync(input, "data")).toThrow(Error);
    });

    it("path traversal with same start", () => {
        const input = "../path-other";
        expect(() => file.writeFileSync(input, "data")).toThrow(Error);
    });

    it("safe path traversal", () => {
        const input = "path";
        file.writeFileSync(input, "data");
    });
});

describe("readFileSync", () => {
    it("path traversal", () => {
        const input = "../path";
        expect(() => file.readFileSync(input)).toThrow(Error);
    });

    it("path traversal with trailing", () => {
        const input = "../path/";
        expect(() => file.readFileSync(input)).toThrow(Error);
    });

    it("path traversal with same start", () => {
        const input = "../path-other";
        expect(() => file.readFileSync(input)).toThrow(Error);
    });

    it("safe path traversal", () => {
        const input = "path";
        file.readFileSync(input);
    });
});

describe("mkdirSync", () => {
    it("path traversal", () => {
        const input = "../path";
        expect(() => file.mkdirSync(input)).toThrow(Error);
    });

    it("path traversal with trailing", () => {
        const input = "../path/";
        expect(() => file.mkdirSync(input)).toThrow(Error);
    });

    it("path traversal with same start", () => {
        const input = "../path-other";
        expect(() => file.mkdirSync(input)).toThrow(Error);
    });

    it("safe path traversal", () => {
        const input = "path";
        file.mkdirSync(input);
    });
});
