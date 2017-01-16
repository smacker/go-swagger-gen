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

func assertOperationParamsPost(t *testing.T, op *spec.Operation) {
	assert.NotNil(t, op)
	assert.Len(t, op.Parameters, 1)

	param := op.Parameters[0]
	assert.Equal(t, "content", param.Name)
	assert.Equal(t, "body", param.In)
	assert.Equal(t, "object", param.Type)
	assert.Equal(t, "", param.Format)
	assert.True(t, param.Required)
	assert.Equal(t, "Order", param.Description)
	assert.Equal(t, "#/definitions/order", param.Schema.Ref.String())
}
