package scan

import (
	goparser "go/parser"
	"log"
	"testing"

	"github.com/go-openapi/spec"
	"github.com/stretchr/testify/assert"
)

func TestRoutesParserParameters(t *testing.T) {
	docFile := "../fixtures/goparsing/classification/operations/todo_operation.go"
	fileTree, err := goparser.ParseFile(classificationProg.Fset, docFile, nil, goparser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

	rp := newRoutesParser(classificationProg)
	var ops spec.Paths
	err = rp.Parse(fileTree, &ops)
	assert.NoError(t, err)

	assert.Len(t, ops.Paths, 3)

	po, ok := ops.Paths["/orders/{id}"]
	assert.True(t, ok)
	assert.NotNil(t, po.Get)

	assertOperationParamsGet(t, po.Get)
	assertOperationResponseGet(t, po.Get)

	po, ok = ops.Paths["/orders"]
	assert.NotNil(t, po.Post)
	assertOperationParamsPost(t, po.Post)
}

func assertOperationParamsGet(t *testing.T, op *spec.Operation) {
	assert.NotNil(t, op)
	assert.Len(t, op.Parameters, 2)

	param := op.Parameters[0]
	assert.Equal(t, "id", param.Name)
	assert.Equal(t, "path", param.In)
	assert.Equal(t, "int", param.Type)
	assert.Equal(t, "", param.Format)
	assert.True(t, param.Required)
	assert.Equal(t, "Order ID", param.Description)

	param = op.Parameters[1]
	assert.Equal(t, "name", param.Name)
	assert.Equal(t, "query", param.In)
	assert.Equal(t, "string", param.Type)
	assert.Equal(t, "uuid", param.Format)
	assert.False(t, param.Required)
	assert.Equal(t, "Filter by name", param.Description)
}

func assertOperationResponseGet(t *testing.T, op *spec.Operation) {
	assert.NotNil(t, op.Responses.Default)
	assert.Equal(t, "#/responses/genericError", op.Responses.Default.Ref.String())

	responses := op.Responses.ResponsesProps.StatusCodeResponses
	assert.Len(t, responses, 3)

	// annotated response
	rsp, ok := responses[200]
	assert.True(t, ok)

	assert.Equal(t, []string{"name"}, rsp.Schema.Required)
	assert.Equal(t, "OK", rsp.Description)
	assertProperty(t, rsp.Schema, "integer", "count", "", "")
	assertProperty(t, rsp.Schema, "string", "name", "email", "")

	rsp, ok = responses[403]
	assert.True(t, ok)
	assert.Equal(t, "Access denided", rsp.Description)
	assert.Equal(t, "#/definitions/validationError", rsp.Schema.Ref.String())

	// original response
	rsp, ok = responses[422]
	assert.True(t, ok)
	assert.Equal(t, "#/responses/validationError", rsp.Ref.String())
}

func assertOperationParamsPost(t *testing.T, op *spec.Operation) {
	assert.NotNil(t, op)
	assert.Len(t, op.Parameters, 2)

	param := op.Parameters[0]
	assert.Equal(t, "content", param.Name)
	assert.Equal(t, "body", param.In)
	assert.Equal(t, "", param.Type)
	assert.Equal(t, "", param.Format)
	assert.True(t, param.Required)
	assert.Equal(t, "Order", param.Description)
	assert.Equal(t, "#/definitions/order", param.Schema.Ref.String())

	param = op.Parameters[1]
	assert.Equal(t, "someType", param.Name)
	assert.Equal(t, "query", param.In)
	assert.Equal(t, "string", param.Type)
	assert.Equal(t, "", param.Format)
	assert.True(t, param.Required)
	assert.Equal(t, "type of content", param.Description)
	assert.Equal(t, []interface{}{
		"value1", "value2", "value3",
	}, param.Enum)
}
