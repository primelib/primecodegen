package openapidocument

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"

	"github.com/pb33f/libopenapi/datamodel/high/base"
)

func HashSchema(schema *base.SchemaProxy) string {
	b, _ := json.Marshal(schema.Schema())
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])[:6]
}

func Compare(aProxy *base.SchemaProxy, bProxy *base.SchemaProxy) bool {
	return CompareWithVisited(aProxy, bProxy, make(map[schemaPair]bool))
}

type schemaPair struct {
	a, b *base.SchemaProxy
}

// CompareWithVisited compares schemas and avoids infinite loops
func CompareWithVisited(aProxy, bProxy *base.SchemaProxy, visited map[schemaPair]bool) bool {
	if aProxy == nil || bProxy == nil {
		return aProxy == bProxy
	}

	pair := schemaPair{aProxy, bProxy}
	if visited[pair] {
		// Already compared these exact proxies — avoid infinite loop
		return true
	}
	visited[pair] = true

	a := aProxy.Schema()
	b := bProxy.Schema()
	if a == nil || b == nil {
		return a == b
	}

	// basic
	if !equalStringSlices(a.Type, b.Type) {
		return false
	}
	if !equalStringSlices(a.Required, b.Required) {
		return false
	}
	if a.Format != b.Format ||
		a.Description != b.Description ||
		a.Nullable != b.Nullable ||
		a.ReadOnly != b.ReadOnly ||
		a.WriteOnly != b.WriteOnly {
		return false
	}

	// properties
	if a.Properties == nil || b.Properties == nil {
		return a.Properties == b.Properties
	}
	if a.Properties.Len() != b.Properties.Len() {
		return false
	}
	for p := a.Properties.Oldest(); p != nil; p = p.Next() {
		propB, ok := b.Properties.Get(p.Key)
		if !ok {
			return false
		}
		if !CompareWithVisited(p.Value, propB, visited) {
			return false
		}
	}

	// TODO: Compare Items, AllOf, OneOf, AnyOf, constraints, etc.

	return true
}

func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	as := append([]string(nil), a...)
	bs := append([]string(nil), b...)
	sort.Strings(as)
	sort.Strings(bs)
	for i := range as {
		if as[i] != bs[i] {
			return false
		}
	}
	return true
}
