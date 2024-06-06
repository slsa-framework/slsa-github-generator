/*
Copyright 2022 SLSA Authors
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

import * as reopener from "./reopener";
import * as config from "./config";

export async function runAction(): Promise<void> {
  const verifiedToken = core.getInput("slsa-verified-token", {
    required: true,
  });
  const outputPath = core.getInput("output-predicate", { required: true });
  const builderInterfaceType = core.getInput("builder-interface-type", {
    required: true,
  });

  // TODO: Implement
}
