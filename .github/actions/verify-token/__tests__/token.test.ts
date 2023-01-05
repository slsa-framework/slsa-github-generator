import { extractIdentityFromSAN } from "../src/token";

describe("extractIdentityFromSAN", () => {
  it("extracts repo and ref from URI", () => {
    const uri =
      "https://github.com/username/reponame/path/to/workflow.yml@refs/heads/main";
    const [repoName, ref] = extractIdentityFromSAN(uri);
    expect(repoName).toEqual("username/reponame");
    expect(ref).toEqual("refs/heads/main");
  });

  it("validates GitHub URL", () => {
    const uri =
      "https://example.com/username/reponame/path/to/workflow.yml@refs/heads/main";
    try {
      extractIdentityFromSAN(uri);
    } catch (e) {
      return;
    }

    fail("expected error");
  });
});
