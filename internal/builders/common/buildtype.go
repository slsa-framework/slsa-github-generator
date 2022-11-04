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
