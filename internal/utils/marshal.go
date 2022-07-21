// Copyright 2022 SLSA Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package utils

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
)

// UnmarshalList unmarshals a string into a list of strings.
func UnmarshalList(arg string) ([]string, error) {
	var res []string
	// If argument is empty, return an empty list early,
	// because `json.Unmarshal` would fail.
	if arg == "" {
		return res, nil
	}

	cs, err := base64.StdEncoding.DecodeString(arg)
	if err != nil {
		return res, fmt.Errorf("base64.StdEncoding.DecodeString: %w", err)
	}

	if err := json.Unmarshal(cs, &res); err != nil {
		return []string{}, fmt.Errorf("json.Unmarshal: %w", err)
	}
	return res, nil
}

// MarshalToString marshals to a string.
func MarshalToString(args interface{}) (string, error) {
	jsonData, err := json.Marshal(args)
	if err != nil {
		return "", fmt.Errorf("json.Marshal: %w", err)
	}

	encoded := base64.StdEncoding.EncodeToString(jsonData)
	if err != nil {
		return "", fmt.Errorf("base64.StdEncoding.EncodeString: %w", err)
	}
	return encoded, nil
}

// MarshalToBytes marshals to a byte array.
func MarshalToBytes(args interface{}) ([]byte, error) {
	encoded, err := MarshalToString(args)
	if err != nil {
		return nil, err
	}
	return []byte(encoded), nil
}
