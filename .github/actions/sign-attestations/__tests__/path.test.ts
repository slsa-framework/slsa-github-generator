import { resolvePathInput } from "../src/path";

// Import using commonJS so it can be mocked.
const fs = require("fs");

jest.mock("fs");

describe("resolvePathInput", () => {
  it("resolves a path", () => {
    const path = resolvePathInput("test", "/working-dir");
    expect(path).toEqual("/working-dir/test");
  });

  it("resolves a path with ..", () => {
    const path = resolvePathInput("test/../foo", "/working-dir");
    expect(path).toEqual("/working-dir/foo");
  });

  it("raises an error for invalid path", () => {
    try {
      resolvePathInput("test/../../foo", "/working-dir");
    } catch (e) {
      return;
    }

    fail("expected error");
  });
});
