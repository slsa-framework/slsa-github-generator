const utils = require("../src/utils");

describe("resolvePathInput", () => {
  it("path traversal", async () => {
    const input = "../path";
    const wd = "/home";
    await expect(
      async () => await utils.resolvePathInput(input, wd)
    ).rejects.toThrow(Error);
  });

  it("path traversal with trailing", async () => {
    const input = "../path";
    const wd = "/home/";
    await expect(
      async () => await utils.resolvePathInput(input, wd)
    ).rejects.toThrow(Error);
  });

  it("path traversal with same start", async () => {
    const input = "../path-other";
    const wd = "/home/path";
    await expect(
      async () => await utils.resolvePathInput(input, wd)
    ).rejects.toThrow(Error);
  });

  it("safe path traversal", async () => {
    const input = "path";
    const wd = "/home";

    const safe = await utils.resolvePathInput(input, wd);
    expect(safe).toEqual("/home/path");
  });
});
