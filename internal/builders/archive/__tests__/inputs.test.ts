// Copyright 2023 SLSA Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

/**
 * @fileoverview Tests for inputs.ts
 */

import * as inputs from "../src/inputs";

describe("parseFormats", () => {
  it("empty value", async () => {
    const value = "";
    expect(() => {
      const formats = inputs.parseFormats(value);
    }).toThrow();
  });

  it("space value", async () => {
    const value = "   ";
    expect(() => {
      const formats = inputs.parseFormats(value);
    }).toThrow();
  });

  it("eol value", async () => {
    const value = `
    `;
    expect(() => {
      const formats = inputs.parseFormats(value);
    }).toThrow();
  });

  it("eol + space value", async () => {
    const value = `
       
    `;
    expect(() => {
      const formats = inputs.parseFormats(value);
    }).toThrow();
  });

  it("one line one format", async () => {
    const value = "zip";
    const formats = inputs.parseFormats(value);
    expect(formats).toEqual(["zip"]);
  });

  it("one line one format with space", async () => {
    const value = " zip ";
    const formats = inputs.parseFormats(value);
    expect(formats).toEqual(["zip"]);
  });

  it("one line two format", async () => {
    const value = "zip tar.gz";
    const formats = inputs.parseFormats(value);
    expect(formats).toEqual(["zip", "tar.gz"]);
  });

  it("one line two format with space", async () => {
    const value = " zip tar.gz  ";
    const formats = inputs.parseFormats(value);
    expect(formats).toEqual(["zip", "tar.gz"]);
  });

  it("two line two format with tab", async () => {
    const value = `zip
                  tar.gz`;
    const formats = inputs.parseFormats(value);
    expect(formats).toEqual(["zip", "tar.gz"]);
  });

  it("two line two format with tab and space", async () => {
    const value = `  zip  
                     tar.gz  `;
    const formats = inputs.parseFormats(value);
    expect(formats).toEqual(["zip", "tar.gz"]);
  });
});

describe("formatsToAPI", () => {
  it("empty formats", async () => {
    const formats: string[] = [];
    expect(() => {
      const actual = inputs.formatsToAPI(formats);
    }).toThrow();
  });

  it("two valid formats", async () => {
    const formats = ["zip", "tar.gz"];
    const actual = inputs.formatsToAPI(formats);
    expect(actual).toEqual(["zipball", "tarball"]);
  });

  it("tarball format", async () => {
    const formats = ["tar.gz"];
    const actual = inputs.formatsToAPI(formats);
    expect(actual).toEqual(["tarball"]);
  });

  it("zip format", async () => {
    const formats = ["zip"];
    const actual = inputs.formatsToAPI(formats);
    expect(actual).toEqual(["zipball"]);
  });
});
