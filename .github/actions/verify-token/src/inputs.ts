/*
Copyright 2023 SLSA Authors
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
import * as fetch from 'node-fetch';
import { rawTokenInterface } from "../src/types";


export function fixInputs(slsaToken: rawTokenInterface, 
    ghToken: string, repoName: string, hash: string): rawTokenInterface {
    const ret = Object.create(slsaToken);
    const modifiedInputs = new Map(Object.entries(slsaToken.tool.inputs));

    // 1.3.6.1.4.1.57264.1.3: 
    // 8cbf4d422367d8499d5980a837cb9cc8e1e67001

    // for (const key of token.tool.masked_inputs) {
    //     if (!maskedMapInputs.has(key)) {
    //     throw new Error(`input ${key} does not exist in the input map`);
    //     }
    //     // verify non-empty keys.
    //     if (key === undefined || key.trim().length === 0) {
    //     throw new Error("empty key in the input map");
    //     }
    //     // NOTE: This mask is the same used by GitHub for encrypted secrets and masked values.
    //     maskedMapInputs.set(key, "***");
    // }
    // ret.tool.inputs = maskedMapInputs;
    //
    const headers = new fetch.Headers();
    headers.append("Authorization:", `token ${ghToken}`);
    const response = fetch.default('https://api.github.com/users/github',{headers: headers});
    core.info(`repsonse: ${response}`)
    return ret;
}

// r.Header.Add("Authorization", fmt.Sprintf("Bearer %s", transport.token))
// https://github.com/ossf/scorecard-webapp/blob/main/app/server/github_transport.go