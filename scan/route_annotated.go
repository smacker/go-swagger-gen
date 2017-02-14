package scan

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-openapi/loads/fmts"
	"github.com/go-openapi/spec"
	"gopkg.in/yaml.v2"
)

func newSetParameters(definitions map[string]spec.Schema, setter func([]spec.Parameter)) *setParameters {
	return &setParameters{
		set:         setter,
		rx:          rxParameters,
		definitions: definitions,
	}
}

type setParameters struct {
	set         func([]spec.Parameter)
	rx          *regexp.Regexp
	definitions map[string]spec.Schema
}

func (ss *setParameters) Matches(line string) bool {
	return ss.rx.MatchString(line)
}

// id int in:path required "Order ID"
var paramShortcutRe = regexp.MustCompile(`([\w]+)[\s]*([\S.]+)[\s]+in:([\w]+)[\s]+([\w]+)[\s]+"([^"]+)"`)
var paramYamlRe = regexp.MustCompile("^|")

var yamlResponseRe = regexp.MustCompile(`(\d+):[\s*]"(.+)"`)
var normalReponseRe = regexp.MustCompile(`(\d|default):.+`)

func (ss *setParameters) Parse(lines []string) error {
	if len(lines) == 0 || (len(lines) == 1 && len(lines[0]) == 0) {
		return nil
	}

	result := []spec.Parameter{}
	yamlLines := []string{}
	for _, line := range lines {
		matches := paramShortcutRe.FindStringSubmatch(line)
		if len(matches) == 6 {
			param, err := parseParamShortcut(matches)
			if err != nil {
				return err
			}
			result = append(result, *param)
			continue
		}
		if paramYamlRe.MatchString(line) {
			trimmedLine := strings.Replace(line, "| ", "", 1)
			if trimmedLine != "" {
				yamlLines = append(yamlLines, trimmedLine)
			}
			continue
		}

		return fmt.Errorf("can not parse param comment \"%s\"", line)
	}

	yamlParams, err := parseParamYaml(yamlLines)
	if err != nil {
		return err
	}
	result = append(result, yamlParams...)

	ss.set(result)
	return nil
}

func parseParamShortcut(matches []string) (*spec.Parameter, error) {
	name := matches[1]
	pType := matches[2]
	location := matches[3]
	required := matches[4]
	desc := matches[5]
	format := ""

	typeAndFormat := strings.Split(pType, ":")
	if len(typeAndFormat) == 2 {
		pType = typeAndFormat[0]
		format = typeAndFormat[1]
	}

	var param *spec.Parameter
	switch location {
	case "query":
		param = spec.QueryParam(name)
	case "header":
		param = spec.HeaderParam(name)
	case "path":
		param = spec.PathParam(name)
	case "body":
		ref, err := spec.NewRef("#/definitions/" + pType)
		if err != nil {
			return param, err
		}
		schema := &spec.Schema{}
		schema.Ref = ref
		param = spec.BodyParam(name, schema)
		param.Type = ""
	case "formData":
		param = spec.FormDataParam(name)
	default:
		return param, fmt.Errorf("unknown param location \"%s\"", location)
	}

	switch required {
	case "true", "required":
		param = param.AsRequired()
	case "false":
		param = param.AsOptional()
	default:
		return param, fmt.Errorf("unknown param required string \"%s\"", required)
	}

	if desc != "" {
		param = param.WithDescription(desc)
	}

	if location != "body" {
		param = param.Typed(pType, format)
	}

	return param, nil
}

func parseParamYaml(yamlLines []string) ([]spec.Parameter, error) {
	if len(yamlLines) == 0 {
		return []spec.Parameter{}, nil
	}

	yamlContent := strings.Join(yamlLines, "\n")

	// get yaml value
	yamlValue := []interface{}{}
	if err := yaml.Unmarshal([]byte(yamlContent), &yamlValue); err != nil {
		return nil, err
	}
	result := []spec.Parameter{}
	for _, yamlParam := range yamlValue {
		// convert to json
		var jsonValue json.RawMessage
		jsonValue, err := fmts.YAMLToJSON(yamlParam)
		if err != nil {
			return nil, err
		}
		param := spec.Parameter{}
		if err := param.UnmarshalJSON(jsonValue); err != nil {
			return []spec.Parameter{}, err
		}
		result = append(result, param)
	}

	return result, nil
}

func newSetAnnotatedResponses(sr *setOpResponses) *setAnnotatedResponses {
	return &setAnnotatedResponses{
		originOpResParser: sr,
	}
}

type setAnnotatedResponses struct {
	originOpResParser *setOpResponses
}

func (sar *setAnnotatedResponses) Matches(line string) bool {
	return sar.originOpResParser.Matches(line)
}

func (self *setAnnotatedResponses) Parse(lines []string) error {
	if len(lines) == 0 || (len(lines) == 1 && len(lines[0]) == 0) {
		return nil
	}

	var def *spec.Response
	responsesMap := map[int]*spec.Response{}

	var currentResponse *spec.Response
	inYamlSection := false
	skippedLines := []string{}
	yamlLines := []string{}

	for _, line := range lines {
		yamlMatches := yamlResponseRe.FindStringSubmatch(line)

		if len(yamlMatches) == 3 {
			if err := parseYamlResponse(currentResponse, yamlLines); err != nil {
				return err
			}
			yamlLines = []string{}
			currentResponse = new(spec.Response)
			if sc, err := strconv.Atoi(yamlMatches[1]); err == nil {
				if responsesMap == nil {
					responsesMap = make(map[int]*spec.Response)
				}
				responsesMap[sc] = currentResponse
			}
			currentResponse.Description = yamlMatches[2]
			inYamlSection = true
			continue
		}

		if normalReponseRe.MatchString(line) {
			if err := parseYamlResponse(currentResponse, yamlLines); err != nil {
				return err
			}
			yamlLines = []string{}
			inYamlSection = false
		}

		if inYamlSection {
			trimmedLine := strings.Replace(line, "| ", "", 1)
			if trimmedLine != "" {
				yamlLines = append(yamlLines, trimmedLine)
			}
		} else {
			skippedLines = append(skippedLines, line)
		}
	}

	if inYamlSection {
		if err := parseYamlResponse(currentResponse, yamlLines); err != nil {
			return err
		}
	}

	setter := func(originResponseDef *spec.Response, originResponseScr map[int]spec.Response) {
		def = originResponseDef

		for responseCode := range originResponseScr {
			if _, ok := responsesMap[responseCode]; !ok {
				resp := originResponseScr[responseCode]
				responsesMap[responseCode] = &resp
			}
		}
	}

	originalSetter := self.originOpResParser.set
	self.originOpResParser.set = setter

	if err := self.originOpResParser.Parse(skippedLines); err != nil {
		return err
	}

	// convert map[int]*spec.Response to map[int]spec.Response
	scrValues := make(map[int]spec.Response)
	for code, res := range responsesMap {
		scrValues[code] = *res
	}

	originalSetter(def, scrValues)

	return nil
}

func parseYamlResponse(resp *spec.Response, yamlLines []string) error {
	if resp == nil {
		return nil
	}

	if len(yamlLines) > 0 {
		yamlContent := strings.Join(yamlLines, "\n")

		// get yaml value
		yamlValue := make(map[interface{}]interface{})
		if err := yaml.Unmarshal([]byte(yamlContent), &yamlValue); err != nil {
			return err
		}
		// convert to json
		var jsonValue json.RawMessage
		jsonValue, err := fmts.YAMLToJSON(yamlValue)
		if err != nil {
			return err
		}
		// put json in response
		if err := resp.UnmarshalJSON(jsonValue); err != nil {
			return err
		}
	}

	return nil
}
