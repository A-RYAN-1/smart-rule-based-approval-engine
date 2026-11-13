// Package speccheck compares the OpenAPI description with the routes a gin
// router has registered. It is imported by tests only, so the YAML parser
// never reaches the server binary.
package speccheck

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"go.yaml.in/yaml/v3"
)

var httpMethods = map[string]bool{
	"get": true, "post": true, "put": true, "delete": true, "patch": true, "head": true, "options": true,
}

var ginParam = regexp.MustCompile(`:([A-Za-z_]+)`)

// Operations returns every "METHOD /path" the document describes.
func Operations(spec []byte) ([]string, error) {
	var doc struct {
		Paths map[string]map[string]yaml.Node `yaml:"paths"`
	}
	if err := yaml.Unmarshal(spec, &doc); err != nil {
		return nil, fmt.Errorf("parse spec: %w", err)
	}
	var ops []string
	for path, item := range doc.Paths {
		for key := range item {
			if httpMethods[key] {
				ops = append(ops, strings.ToUpper(key)+" "+path)
			}
		}
	}
	sort.Strings(ops)
	return ops, nil
}

// FromGin turns gin's "METHOD /api/leaves/:id/cancel" into the OpenAPI
// spelling "METHOD /api/leaves/{id}/cancel".
func FromGin(method, path string) string {
	return method + " " + ginParam.ReplaceAllString(path, "{$1}")
}

// Diff lists operations served but not described (undocumented) and
// described but not served (unserved). Entries in ignore are dropped from
// both lists.
func Diff(described, served, ignore []string) (undocumented, unserved []string) {
	skip := toSet(ignore)
	d, s := toSet(described), toSet(served)
	for op := range s {
		if !d[op] && !skip[op] {
			undocumented = append(undocumented, op)
		}
	}
	for op := range d {
		if !s[op] && !skip[op] {
			unserved = append(unserved, op)
		}
	}
	sort.Strings(undocumented)
	sort.Strings(unserved)
	return undocumented, unserved
}

func toSet(in []string) map[string]bool {
	m := make(map[string]bool, len(in))
	for _, v := range in {
		m[v] = true
	}
	return m
}
