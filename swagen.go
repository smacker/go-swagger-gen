// Copyright 2015 go-swagger maintainers
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	swaggererrors "github.com/go-openapi/errors"
	"github.com/go-openapi/loads"
	"github.com/go-openapi/spec"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/validate"
	"github.com/jessevdk/go-flags"
	"github.com/smacker/go-swagger-gen/clean"
	"github.com/smacker/go-swagger-gen/scan"
)

var opts struct{}

func main() {
	parser := flags.NewParser(&opts, flags.Default)
	if _, err := parser.AddCommand("spec", "generate spec", "generate spec file from go code", &SpecFile{}); err != nil {
		log.Fatal(err)
	}
	if _, err := parser.AddCommand("validate", "validate the swagger document", "validate the provided swagger document against a swagger spec", &ValidateSpec{}); err != nil {
		log.Fatal(err)
	}
	if _, err := parser.Parse(); err != nil {
		os.Exit(1)
	}
}

// SpecFile command to generate a swagger spec from a go application
type SpecFile struct {
	BasePath   string         `long:"base-path" short:"b" description:"the base path to use" default:"."`
	ScanModels bool           `long:"scan-models" short:"m" description:"includes models that were annotated with 'swagger:model'"`
	Compact    bool           `long:"compact" description:"when present, doesn't prettify the the json"`
	Output     flags.Filename `long:"output" short:"o" description:"the file to write to"`
	Input      flags.Filename `long:"input" short:"i" description:"the file to use as input"`
}

// Execute runs this command
func (s *SpecFile) Execute(args []string) error {
	input, err := loadSpec(string(s.Input))
	if err != nil {
		return err
	}

	var opts scan.Opts
	opts.BasePath = s.BasePath
	opts.Input = input
	opts.ScanModels = s.ScanModels
	swspec, err := scan.Application(opts)
	if err != nil {
		return err
	}

	clean.RemoveUnusedDefinitions(swspec)

	return writeToFile(swspec, !s.Compact, string(s.Output))
}

var (
	newLine = []byte("\n")
)

func loadSpec(input string) (*spec.Swagger, error) {
	if fi, err := os.Stat(input); err == nil {
		if fi.IsDir() {
			return nil, fmt.Errorf("expected %q to be a file not a directory", input)
		}
		sp, err := loads.Spec(input)
		if err != nil {
			return nil, err
		}
		return sp.Spec(), nil
	}
	return nil, nil
}

func writeToFile(swspec *spec.Swagger, pretty bool, output string) error {
	var b []byte
	var err error
	if pretty {
		b, err = json.MarshalIndent(swspec, "", "  ")
	} else {
		b, err = json.Marshal(swspec)
	}
	if err != nil {
		return err
	}
	if output == "" {
		fmt.Println(string(b))
		return nil
	}
	return ioutil.WriteFile(output, b, 0644)
}

// ValidateSpec is a command that validates a swagger document
// against the swagger json schema
type ValidateSpec struct {
}

// Execute validates the spec
func (c *ValidateSpec) Execute(args []string) error {
	if len(args) == 0 {
		return errors.New("The validate command requires the swagger document url to be specified")
	}

	swaggerDoc := args[0]
	specDoc, err := loads.Spec(swaggerDoc)
	if err != nil {
		log.Fatalln(err)
	}

	result := validate.Spec(specDoc, strfmt.Default)
	if result == nil {
		fmt.Printf("The swagger spec at %q is valid against swagger specification %s\n", swaggerDoc, specDoc.Version())
	} else {
		str := fmt.Sprintf("The swagger spec at %q is invalid against swagger specification %s. see errors :\n", swaggerDoc, specDoc.Version())
		for _, desc := range result.(*swaggererrors.CompositeError).Errors {
			str += fmt.Sprintf("- %s\n", desc)
		}
		return errors.New(str)
	}
	return nil
}
