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

package common

import "github.com/slsa-framework/slsa-github-generator/slsa"

// GenericBuild is a very generic build type where build type can be specified.
type GenericBuild struct {
	*slsa.GithubActionsBuild
	BuildTypeURI string
}

// URI implements BuildType.URI.
func (b *GenericBuild) URI() string {
	return b.BuildTypeURI
}
