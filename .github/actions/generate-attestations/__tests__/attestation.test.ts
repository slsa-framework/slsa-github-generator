import * as types from "../src/intoto";
import * as statement from "../src/attestation";

// Import using commonJS so it can be mocked.
const fs = require("fs");

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
    expect(attestation._type).toEqual(types.INTOTO_TYPE);
  });
});

jest.mock("fs");

describe("writeAttestation", () => {
  afterEach(() => {
    jest.clearAllMocks();
  });

  const layout = {
    version: 1,
    attestations: [
      {
        name: "attestation1.intoto",
        subjects: [
          {
            name: "first",
            digest: { sha256: "deadbeaf" },
          },
        ],
      },
      {
        name: "attestation2.intoto",
        subjects: [
          {
            name: "second",
            digest: { sha256: "deadbeaf" },
          },
        ],
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

  it("writes correct intoto statement", () => {
    fs.readFileSync.mockReturnValueOnce(JSON.stringify(layout));
    fs.readFileSync.mockReturnValueOnce(JSON.stringify(predicate));

    const attestations: Record<string, string> = statement.writeAttestations(
      "layoutMock",
      predicateType,
      "predicateMock"
    );
    expect(Object.keys(attestations)).toEqual(
      expect.arrayContaining(["attestation1.intoto", "attestation2.intoto"])
    );
  });

  it("nested intoto statement names", () => {
    const nestedLayout = {
      version: 1,
      attestations: [
        {
          name: "attestation1.intoto",
          subjects: [
            {
              name: "first",
              digest: { sha256: "deadbeaf" },
            },
          ],
        },
        {
          name: "bad/attestation2.intoto",
          subjects: [
            {
              name: "second",
              digest: { sha256: "deadbeaf" },
            },
          ],
        },
      ],
    };

    fs.readFileSync.mockReturnValueOnce(JSON.stringify(nestedLayout));
    fs.readFileSync.mockReturnValueOnce(JSON.stringify(predicate));

    expect(() =>
      statement.writeAttestations("layoutMock", predicateType, "predicateMock")
    ).toThrow(Error);
  });

  it("too many attestations", () => {
    for (var i = 0; i < 51; i++) {
      layout.attestations.push({
        name: i.toString(),
        subjects: [{ name: "first", digest: { sha256: "deadbeef" } }],
      });
    }

    fs.readFileSync.mockReturnValueOnce(JSON.stringify(layout));
    fs.readFileSync.mockReturnValueOnce(JSON.stringify(predicate));

    expect(() =>
      statement.writeAttestations("layoutMock", predicateType, "predicateMock")
    ).toThrow(Error);
  });
});
