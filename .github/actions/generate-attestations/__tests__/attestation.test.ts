const fs = require("fs");
const statement = require("../src/attestation");
const intoto = require("../src/intoto");

describe("createStatement", () => {
  const subjects = [
    {
      name: "result",
      digest: { sha256: "deadbeaf" },
    },
  ];
  const predicate = {
    buildType: "https://github.com/slsa-framework/slsa-github-generator/go@v1",
    builder: {
      id: "https://github.com/slsa-framework/slsa-github-generator/.github/workflows/builder_go_slsa3.yml@refs/tags/v0.0.1",
    },
  };

  const predicateType = "https://slsa.dev/provenance/v0.2";

  it("returns the correct intoto statement", () => {
    const attestation = statement.createStatement(
      subjects,
      predicateType,
      predicate
    );
    expect(attestation.subject).toEqual(subjects);
    expect(attestation.predicateType).toEqual(predicateType);
    expect(attestation._type).toEqual(intoto.INTOTO_TYPE);
  });
});

jest.mock("fs");

describe("writeAttestation", () => {
  const layout = {
    version: 1,
    "attestation1.intoto": [
      {
        name: "first",
        digest: { sha256: "deadbeaf" },
      },
    ],
    "attestation2.intoto": [
      {
        name: "second",
        digest: { sha256: "deadc0de" },
      },
    ],
  };

  const predicate = {
    buildType: "https://github.com/slsa-framework/slsa-github-generator/go@v1",
    builder: {
      id: "https://github.com/slsa-framework/slsa-github-generator/.github/workflows/builder_go_slsa3.yml@refs/tags/v0.0.1",
    },
  };

  const predicateType = "https://slsa.dev/provenance/v0.2";

  it("writes correct intoto statement", async () => {
    fs.readFileSync.mockClear();

    fs.readFileSync.mockReturnValueOnce(JSON.stringify(layout));
    fs.readFileSync.mockReturnValueOnce(JSON.stringify(predicate));

    const attestations: Record<string, string> =
      await statement.writeAttestations(
        "layoutMock",
        predicateType,
        "predicateMock"
      );
    expect(Object.keys(attestations)).toEqual(
      expect.arrayContaining(["attestation1.intoto", "attestation2.intoto"])
    );
  });
});
