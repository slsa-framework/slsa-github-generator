package main

import (
	"fmt"
	"strings"

	"github.com/pborman/uuid"
)

var (
	gitVersion string
	gitCommit  string
	gitBranch  string
)

func main() {
	uuidWithHyphen := uuid.NewRandom()
	uuid := strings.Replace(uuidWithHyphen.String(), "-", "", -1)

	fmt.Println("GitBranch:", gitBranch)
	fmt.Println("GitVersion:", gitVersion)
	fmt.Println("GitCommit:", gitCommit)
	fmt.Println("Hello world:", uuid)
}
