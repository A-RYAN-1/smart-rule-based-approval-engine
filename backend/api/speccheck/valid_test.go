package speccheck_test

import (
	"context"
	"testing"

	apidocs "github.com/ankita-advitot/rule_based_approval_engine/api"
	"github.com/getkin/kin-openapi/openapi3"
)

// TestSpecIsValidOpenAPI loads the embedded document with a real OpenAPI 3
// parser: broken references, bad schemas and malformed paths fail here.
func TestSpecIsValidOpenAPI(t *testing.T) {
	doc, err := openapi3.NewLoader().LoadFromData(apidocs.Spec)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if err := doc.Validate(context.Background()); err != nil {
		t.Fatalf("validate: %v", err)
	}
	if len(doc.Paths.Map()) < 40 {
		t.Fatalf("only %d paths in the spec", len(doc.Paths.Map()))
	}
}
