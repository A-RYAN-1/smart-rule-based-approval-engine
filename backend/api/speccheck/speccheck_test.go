package speccheck

import (
	"reflect"
	"testing"
)

func TestOperationsAndDiff(t *testing.T) {
	spec := []byte(`
openapi: 3.0.3
paths:
  /api/a:
    parameters: []
    get: {}
    post: {}
  /api/a/{id}:
    delete: {}
`)
	got, err := Operations(spec)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"DELETE /api/a/{id}", "GET /api/a", "POST /api/a"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Operations = %v, want %v", got, want)
	}

	served := []string{"GET /api/a", FromGin("DELETE", "/api/a/:id"), "PUT /api/b", "GET /api/docs"}
	undocumented, unserved := Diff(got, served, []string{"GET /api/docs"})
	if !reflect.DeepEqual(undocumented, []string{"PUT /api/b"}) {
		t.Errorf("undocumented = %v", undocumented)
	}
	if !reflect.DeepEqual(unserved, []string{"POST /api/a"}) {
		t.Errorf("unserved = %v", unserved)
	}
}
