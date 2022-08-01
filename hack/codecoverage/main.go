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
package main

import (
	"bufio"
	"encoding/json"
	"log"
	"os"
	"strconv"
	"strings"
)

func main() {
	if len(os.Args) != 3 {
		log.Fatal("Usage: <threshold-file> <coverage-percent>")
	}
	thresholdFile := os.Args[1]
	if thresholdFile == "" {
		log.Fatalf("The thresholdfile cannot be empty.")
	}
	thresholdMap, err := parseCoverageThreshold(thresholdFile)
	if err != nil {
		log.Fatalf("Error parsing threshold file: %v", err)
	}
	coveragePercentage := os.Args[2]
	if coveragePercentage == "" {
		log.Fatalf("The coverage percentage cannot be empty.")
	}
	coveragePercentageFloat, err := strconv.ParseFloat(coveragePercentage, 64)
	if err != nil {
		log.Fatalf("Error parsing coverage percentage: %v", err)
	}
	// read stream from stdin
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "coverage: ") {
			parts := strings.Fields(line)
			if len(parts) < 5 {
				continue
			}
			percentage, err := strconv.ParseFloat(strings.Trim(parts[4], "%"), 64)
			if err != nil {
				log.Fatalf("invalid line: %s", line)
			}
			pack := parts[1]
			if val, ok := thresholdMap[pack]; ok {
				if percentage < val {
					log.Fatalf("coverage for %s is below threshold: %f < %f", pack, percentage, val)
				}
				continue
			}
			// if the package is not in the threshold map, then we check if the coverage is below the threshold
			if percentage < coveragePercentageFloat {
				log.Fatalf("coverage for %s is below threshold: %f < %f", pack, percentage, coveragePercentageFloat)
			}
		}
	}
}

// parseCoverageThreshold parses the threshold file and returns a map.
func parseCoverageThreshold(fileName string) (map[string]float64, error) {
	// Here is an example of the threshold file:
	/*
		{
			  "github.com/foo/bar/pkg/cryptoutils": 71.2,
			}
	*/
	f, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	thresholdMap := make(map[string]float64)
	if err := json.NewDecoder(f).Decode(&thresholdMap); err != nil {
		return nil, err
	}
	return thresholdMap, nil
}
