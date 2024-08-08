package main

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"log"
	"os"
	"strings"
)

type SLSALayout struct {
	Version      int            `json:"version"`
	Attestations []*Attestation `json:"attestations"`
}
type Attestation struct {
	Name     string     `json:"name"`
	Subjects []*Subject `json:"subjects"`
}

type Subject struct {
	Name   string            `json:"name"`
	Digest map[string]string `json:"digest"`
}

func main() {
	base64Subjects := flag.String("base64-subjects", "", "a base64-encoded list of subjects")
	base64SubjectsFile := flag.String("base64-subjects-file", "", "file with a base64-encoded list of subjects")
	provenanceName := flag.String("provenance-name", "", "name of the provenance, including the .build.slsa suffix")
	outputFile := flag.String("output-file", "", "outfile to write the SLSA layout to")
	flag.Parse()

	if !strings.HasSuffix(*provenanceName, ".build.slsa") {
		log.Fatalf("provenance name must have the .build.slsa suffix: %s", *provenanceName)
	}

	var base64Content string
	if *base64Subjects != "" {
		base64Content = *base64Subjects
	} else if *base64SubjectsFile != "" {
		base64ContentBytes, err := os.ReadFile(*base64SubjectsFile)
		if err != nil {
			log.Fatalf("failed to read file %s: %v", *base64SubjectsFile, err)
		}
		base64Content = string(base64ContentBytes)
	} else {
		log.Fatal("either --base64-subjects or --base64-subjects-file must be set")
	}

	decodedContent, err := base64.StdEncoding.DecodeString(base64Content)
	if err != nil {
		log.Fatalf("failed to decode base64 the content. Did you base64-encode?: %v", err)
	}

	attestation := Attestation{
		Name: strings.TrimSuffix(*provenanceName, ".build.slsa"),
	}
	layout := SLSALayout{
		Version:      1,
		Attestations: []*Attestation{&attestation},
	}

	lines := strings.Split(string(decodedContent), "\n")
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		parts := strings.SplitN(line, " ", 2)
		if len(parts) != 2 {
			log.Fatalf("invalid line, should be `<artifact digest> <artifact name>`: %s", line)
		}
		artifactDigest := parts[0]
		artifactName := parts[1]

		subject := &Subject{
			Name: artifactName,
			Digest: map[string]string{
				"sha256": artifactDigest,
			},
		}

		attestation.Subjects = append(attestation.Subjects, subject)
	}
	if len(attestation.Subjects) == 0 {
		log.Fatal("no subjects found")
	}

	layoutBytes, err := json.Marshal(layout)
	if err != nil {
		log.Fatalf("failed to marshal layout: %v", err)
	}

	f, err := os.Create(*outputFile)
	if err != nil {
		log.Fatalf("failed to create file %s: %v", *outputFile, err)
	}
	defer f.Close()

	_, err = f.Write(layoutBytes)
	if err != nil {
		log.Fatalf("failed to write layout to file: %v", err)
	}
	return
}
