package scan

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/go-openapi/spec"
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
var paramRe = regexp.MustCompile(`([\w]+)[\s]*([\S.]+)[\s]+in:([\w]+)[\s]+([\w]+)[\s]+"([^"]+)"`)

func (ss *setParameters) Parse(lines []string) error {
	if len(lines) == 0 || (len(lines) == 1 && len(lines[0]) == 0) {
		return nil
	}

	result := make([]spec.Parameter, len(lines))
	for i, line := range lines {
		matches := paramRe.FindStringSubmatch(line)
		if len(matches) != 6 {
			return fmt.Errorf("can not parse param comment \"%s\"", line)
		}

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
				return err
			}
			schema := &spec.Schema{}
			schema.Ref = ref
			param = spec.BodyParam(name, schema)
		case "formData":
			param = spec.FormDataParam(name)
		default:
			return fmt.Errorf("unknown param location \"%s\"", location)
		}

		switch required {
		case "true", "required":
			param = param.AsRequired()
		case "false":
			param = param.AsOptional()
		default:
			return fmt.Errorf("unknown param required string \"%s\"", required)
		}

		if desc != "" {
			param = param.WithDescription(desc)
		}

		if location != "body" {
			param = param.Typed(pType, format)
		}

		result[i] = *param
	}
	ss.set(result)
	return nil
}
