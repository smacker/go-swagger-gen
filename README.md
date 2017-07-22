Go Swagger Generator
==================

Go Swagger Generator is a tool to parse Golang source files and generate Swagger json.

Tool is based on [go-swagger](https://github.com/go-swagger/go-swagger)
and supports everything that go-swagger supports, but allows to define responses and parameters as yaml.

Why?
----

Go-swagger expects separate structure for every response, but in my case we generate it dynamically in controller.

Example
-------

```go
// swagger:route PUT /packages/{id} Package packagePut
//
// Update package
//
// Parameters:
// id      string     in:path required "Package ID"
// content PutPackage in:body required "Update package params"
// | - name: type
// |   in: query
// |   description: Example parameter defined as yaml
// |   required: true
// |   type: string
// |   enum:
// |     - value1
// |     - value2
//
// Responses:
// 200: "OK"
// | schema:
// |   properties:
// |     data:
// |       $ref: "#/definitions/PackageGet"
// |     meta:
// |       $ref: "#/definitions/Meta"
// 404: "Package not found"
// | schema:
// |   $ref: "#/definitions/ErrorResponse"
// 422: "422 Unprocessable Entity"
// | schema:
// |   $ref: "#/definitions/ErrorResponse"
func PackagePut() http.HandlerFunc {
    ...
}
```

Usage
-----

`go-swagger-gen spec -m -i third_party_swagger.json -b project/cmd/http -o swagger.json --compact`

```
Usage:
  go-swagger-gen [OPTIONS] spec [spec-OPTIONS]

generate spec file from go code

Help Options:
  -h, --help             Show this help message

[spec command options]
      -b, --base-path=   the base path to use (default: .)
      -m, --scan-models  includes models that were annotated with 'swagger:model'
          --compact      when present, doesn't prettify the the json
      -o, --output=      the file to write to
      -i, --input=       the file to use as input
```

Documentation
-------------

Please refer to [go-swagger documentation](https://goswagger.io/generate/spec.html), swagger.json section.

What is new compare to go-swagger
-------------

### Parameters shortcuts

Allows to define route params in short form:

```
// swagger:route ...
//
// Parameters:
// <name>  <type>  in:<place>  <is-required> "<description>"
```

Where:
- `name` - parameter name
- `type` - any go type or struct with `swagger:model` annotation
- `place` - path/query/header/body/formData
- `is-required` - true/false or required (works as true)

### Parameters yaml

```
// swagger:route ...
//
// Parameters:
// | valid swagger parameters yaml as array
```

Just put here normal swagger yaml, but prepend it with `| ` symbol.

### Response yaml

```
// swagger:route ...
//
// Responses:
// <code>: "<description>"
// | valid swagger response yaml
```

Where:
- `code` - integer response code
- `description` - response description

Just put here normal swagger yaml, but prepend it with `| ` symbol.

### Define string format of struct field in comment

```go
type MyModelResponseDTO struct {
    // format: date-time
    CreatedAt string `json:"created_at"`
}
```

### time.Time fields are string.date-time by default

```go
type MyModelInputDTO struct {
    CreatedAt time.Time `json:"created_at"`
}
```

### Get required fields from valid/validation struct tag

```go
type MyModelInputDTO struct {
    //don't need to add "required" comment
    CreatedAt time.Time `json:"created_at" valid:"required"`
}
```

### Skip private comments

```go
type MyModelInputDTO struct {
    //technical comment, shouldn't appear as description in swagger
    // but this comment is used as description
    CreatedAt time.Time `json:"created_at" valid:"required"`
}
```

### Better alias support

```go
type Int64 int64

type MyModel struct {
    // min: 10
    IntField Int64 `json:"intField"`
}
```

Aliases in swagger.json present as native types. Therefore all validation comments work as expected for them.

### x-go-name extension was removed

I don't want to expose internal name in swagger

### Don't put unused definitions in result swagger.json

Scanner creates swagger definitions for embedded structures too, but often we don't need them.

Consider example:

```go
type BaseModel struct {
    ...
}

// swagger:model
type MyModel struct {
    BaseModel
    ...
}

// we use only "#/definitions/MyModel" somewhere
// but BaseModel definition will also appear in swagger json
```

Therefore `go-swagger-gen` analyzes final json and removes unused definitions.
