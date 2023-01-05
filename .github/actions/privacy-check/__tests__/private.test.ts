import { privacyCheck } from "../src/private";

// Import using commonJS so it can be mocked.
const github = require("@actions/github");

jest.mock("@actions/github");

describe("privacyCheck", () => {
  afterEach(() => {
    jest.clearAllMocks();
  });

  it("returns true for public repo", async () => {
    github.getOctokit.mockReturnValueOnce({
      rest: {
        repos: {
          get: function () {
            return Promise.resolve({
              data: {
                private: false,
              },
            });
          },
        },
      },
    });

    const [priv, check] = await privacyCheck("name/repo", "", false);
    expect(priv).toBeFalsy();
    expect(check).toBeTruthy();
  });

  it("returns false for private repo", async () => {
    github.getOctokit.mockReturnValueOnce({
      rest: {
        repos: {
          get: function () {
            return Promise.resolve({
              data: {
                private: true,
              },
            });
          },
        },
      },
    });

    const [priv, check] = await privacyCheck("name/repo", "", false);
    expect(priv).toBeTruthy();
    expect(check).toBeFalsy();
  });

  it("returns true for private repo w/ override", async () => {
    github.getOctokit.mockReturnValueOnce({
      rest: {
        repos: {
          get: function () {
            return Promise.resolve({
              data: {
                private: true,
              },
            });
          },
        },
      },
    });

    const [priv, check] = await privacyCheck("name/repo", "", true);
    expect(priv).toBeTruthy();
    expect(check).toBeTruthy();
  });

  it("returns true for public repo w/ override", async () => {
    github.getOctokit.mockReturnValueOnce({
      rest: {
        repos: {
          get: function () {
            return Promise.resolve({
              data: {
                private: false,
              },
            });
          },
        },
      },
    });

    const [priv, check] = await privacyCheck("name/repo", "", true);
    expect(priv).toBeFalsy();
    expect(check).toBeTruthy();
  });
});
