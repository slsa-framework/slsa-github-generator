import * as sha256 from "../src/sha256";

// Import using commonJS so it can be mocked.
const fs = require("fs");

jest.mock("fs");

describe("sha256.sum", () => {
  afterEach(() => {
    jest.clearAllMocks();
  });

  it("generates a hash from existing file", () => {
    const file_contents = "foo bar";
    const file_contents_hash =
      "fbc1a9f858ea9e177916964bd88c3d37b91a1e84412765e29950777f265c4b75";
    fs.existsSync.mockReturnValueOnce(true);
    fs.readFileSync.mockReturnValueOnce(file_contents);

    const hash = sha256.sum("foo.txt");
    expect(hash).toEqual(file_contents_hash);
  });

  it("raises an error for non-existant file", () => {
    const file_contents = "foo bar";
    const file_contents_hash =
      "fbc1a9f858ea9e177916964bd88c3d37b91a1e84412765e29950777f265c4b75";
    fs.existsSync.mockReturnValueOnce(false);
    fs.readFileSync.mockReturnValueOnce(file_contents);

    try {
      sha256.sum("foo.txt");
    } catch (e) {
      return;
    }

    fail("expected error");
  });
});
