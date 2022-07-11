package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
)

func main() {
	if len(os.Args) < 2 {
		log.Println("Usage: sha256sum <file>")
		panic("missing argument: path to the file to compute the SHA256 hash")
	}

	file := os.Args[1]
	if _, err := os.Stat(file); os.IsNotExist(err) {
		panic(fmt.Sprintf("file not found: %s", file))
	}

	data, err := ioutil.ReadFile(file)
	if err != nil {
		panic(fmt.Sprintf("failed to read file: %s", file))
	}

	hash := sha256.Sum256(data)
	log.Printf("computed sha: %s\n", hex.EncodeToString(hash[:]))

	cmd := exec.Command("/usr/bin/env", "bash", "-c",
		"echo ::set-output name=sha256::"+hex.EncodeToString(hash[:]))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		panic(fmt.Sprintf("failed to set output: %s", err))
	}
}
