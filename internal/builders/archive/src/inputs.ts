/*
Copyright 2024 SLSA Authors
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    https://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WIHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

import * as core from "@actions/core";

export function getEnvVariable(name: string): string {
  const val = process.env[name];
  if (!val) {
    throw new Error(`env variable '${name}' is empty`);
  }
  return val;
}

export function getActionInput(name: string): string {
  const val = core.getInput("slsa-workflow-inputs");
  if (!val) {
    throw new Error(`Action input '${name}' is empty`);
  }
  return val;
}

export function parseFormats(val: string): string[] {
  const formats: string[] = [];
  const parts = val.split(/\r\n|\r|\n/);
  for (const v of parts) {
    const p = v.split(" ");
    for (const pp of p) {
      const s = pp.trim();
      if (s.length === 0) {
        continue;
      }
      formats.push(s);
    }
  }
  if (formats.length === 0) {
    throw new Error(`no format found '${val}'`);
  }
  return formats;
}

export function formatsToAPI(formats: string[]): string[] {
  const apiFormats: string[] = [];
  if (formats.length === 0) {
    throw new Error("empty format");
  }
  for (const format of formats) {
    switch (format) {
      case "zip":
        apiFormats.push("zipball");
        break;

      case "tar.gz":
        apiFormats.push("tarball");
        break;

      default:
        throw new Error(`unsupported format: ${format}`);
    }
  }
  return apiFormats;
}
