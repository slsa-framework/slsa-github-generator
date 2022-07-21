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
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Test_MarshallToString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		variables []string
		expected  string
	}{
		{
			name:      "single arg",
			variables: []string{"--arg"},
			expected:  "WyItLWFyZyJd",
		},
		{
			name: "list args",
			variables: []string{
				"/usr/lib/google-golang/bin/go",
				"build", "-mod=vendor", "-trimpath",
				"-tags=netgo",
				"-ldflags=-X main.gitVersion=v1.2.3 -X main.gitSomething=somthg",
			},
			expected: "WyIvdXNyL2xpYi9nb29nbGUtZ29sYW5nL2Jpbi9nbyIsImJ1aWxkIiwiLW1vZD12ZW5kb3IiLCItdHJpbXBhdGgiLCItdGFncz1uZXRnbyIsIi1sZGZsYWdzPS1YIG1haW4uZ2l0VmVyc2lvbj12MS4yLjMgLVggbWFpbi5naXRTb21ldGhpbmc9c29tdGhnIl0=",
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r, err := MarshallToString(tt.variables)
			if err != nil {
				t.Errorf("marshallToString: %v", err)
			}
			if !cmp.Equal(r, tt.expected) {
				t.Errorf(cmp.Diff(r, tt.expected))
			}
		})
	}
}

func Test_UnmarshallList(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		value    string
		expected []string
	}{
		{
			name:     "single arg",
			expected: []string{"--arg"},
			value:    "WyItLWFyZyJd",
		},
		{
			name:  "invalid",
			value: "blabla",
		},
		{
			name: "list args",
			expected: []string{
				"/usr/lib/google-golang/bin/go",
				"build", "-mod=vendor", "-trimpath",
				"-tags=netgo",
				"-ldflags=-X main.gitVersion=v1.2.3 -X main.gitSomething=somthg",
			},
			value: "WyIvdXNyL2xpYi9nb29nbGUtZ29sYW5nL2Jpbi9nbyIsImJ1aWxkIiwiLW1vZD12ZW5kb3IiLCItdHJpbXBhdGgiLCItdGFncz1uZXRnbyIsIi1sZGZsYWdzPS1YIG1haW4uZ2l0VmVyc2lvbj12MS4yLjMgLVggbWFpbi5naXRTb21ldGhpbmc9c29tdGhnIl0=",
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r, err := UnmarshallList(tt.value)
			if err != nil && len(tt.expected) != 0 {
				t.Errorf("UnmarshallList: %v", err)
			}

			if !cmp.Equal(r, tt.expected) {
				t.Errorf(cmp.Diff(r, tt.expected))
			}
		})
	}
}
