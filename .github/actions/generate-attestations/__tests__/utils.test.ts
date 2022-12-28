const utils = require("../src/utils");

describe("resolvePathInput", () => {
  it("path traversal", () => {
    const input = "../path";
    const wd = "/home";
    expect(() => utils.resolvePathInput(input, wd)).toThrow(Error);
  });

  it("path traversal with trailing", () => {
    const input = "../path";
    const wd = "/home/";
    expect(() => utils.resolvePathInput(input, wd)).toThrow(Error);
  });

  it("path traversal with same start", () => {
    const input = "../path-other";
    const wd = "/home/path";
    expect(() => utils.resolvePathInput(input, wd)).toThrow(Error);
  });

  it("safe path traversal", () => {
    const input = "path";
    const wd = "/home";

    const safe = utils.resolvePathInput(input, wd);
    expect(safe).toEqual("/home/path");
  });
});
