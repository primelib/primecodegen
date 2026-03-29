package openapipatch

import (
	"log/slog"
	"slices"
	"strings"

	"github.com/cidverse/go-ptr"
	"github.com/pb33f/libopenapi"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
	"github.com/primelib/primecodegen/pkg/logging"
	"github.com/primelib/primecodegen/pkg/openapi/openapidocument"
	"github.com/primelib/primecodegen/pkg/util"
)

var FixCommonPatch = BuiltInPatcher{
	Type:                "builtin",
	ID:                  "fix-common",
	Description:         "Fixes various common issues.",
	PatchV3DocumentFunc: FixCommon,
}

// FixCommon fixes various common issues
func FixCommon(doc *libopenapi.DocumentModel[v3.Document], config map[string]interface{}) error {
	for schema := doc.Model.Components.Schemas.Oldest(); schema != nil; schema = schema.Next() {
		s := schema.Value.Schema()
		if s == nil {
			continue
		}
		if s.Properties == nil {
			s.Properties = orderedmap.New[string, *base.SchemaProxy]()
		}

		// move additionalProperties.properties to properties
		if s.AdditionalProperties != nil && s.AdditionalProperties.IsA() {
			a := s.AdditionalProperties.A.Schema()
			if a == nil {
				continue
			}

			if a.Properties != nil {
				for p := a.Properties.Oldest(); p != nil; p = p.Next() {
					if _, exists := s.Properties.Get(p.Key); !exists {
						logging.Trace("moving property from additionalProperties to properties", "schema", schema.Key, "property", p.Key)
						s.Properties.Set(p.Key, p.Value)
					} else {
						slog.Warn("property already exists in properties, skipping", "schema", schema.Key, "property", p.Key)
					}
				}
				s.AdditionalProperties = nil
			}
		}
	}

	err := FixInvalidMaxValue(doc, config)
	if err != nil {
		return err
	}

	err = PruneInvalidPaths(doc, config)
	if err != nil {
		return err
	}

	return nil
}

var FixInvalidMaxValuePatch = BuiltInPatcher{
	Type:                "builtin",
	ID:                  "fix-invalid-max-value",
	Description:         "Fixes integer and long schemas where the maximum value is out of bounds for the type, e.g. max: 9223372036854775807 for long.",
	PatchV3DocumentFunc: FixInvalidMaxValue,
}

// FixInvalidMaxValue fixes integers and longs, where the maximum value is out of bounds for the type
func FixInvalidMaxValue(doc *libopenapi.DocumentModel[v3.Document], config map[string]interface{}) error {
	for schema := doc.Model.Components.Schemas.Oldest(); schema != nil; schema = schema.Next() {
		if schema.Value.Schema().Properties == nil {
			continue
		}

		for p := schema.Value.Schema().Properties.Oldest(); p != nil; p = p.Next() {
			s := p.Value.Schema()
			if slices.Contains(s.Type, "integer") && p.Value.Schema().Maximum != nil {
				if *p.Value.Schema().Maximum > 2147483647 {
					p.Value.Schema().Maximum = ptr.Ptr(float64(2147483647))
					logging.Trace("fixing maximum value for integer", "schema", schema.Key, "property", p.Key)
				}
			}
		}
	}

	return nil
}

var FixOperationTagsPatch = BuiltInPatcher{
	Type:                "builtin",
	ID:                  "fix-operation-tags",
	Description:         "Ensures all operations have at least one tag, and that tags are documented in the document",
	PatchV3DocumentFunc: FixOperationTags,
}

// FixOperationTags ensures all operations have tags, and that tags are documented in the document
func FixOperationTags(doc *libopenapi.DocumentModel[v3.Document], config map[string]interface{}) error {
	documentedTags := make(map[string]bool)
	for _, tag := range doc.Model.Tags {
		documentedTags[tag.Name] = true
	}

	for path := doc.Model.Paths.PathItems.Oldest(); path != nil; path = path.Next() {
		for op := path.Value.GetOperations().Oldest(); op != nil; op = op.Next() {
			if len(op.Value.Tags) == 0 {
				// add default tag, if missing
				logging.Trace("operation is missing tags, adding default tag", "path", strings.ToUpper(op.Key)+" "+path.Key)
				op.Value.Tags = append(op.Value.Tags, "default")
			} else {
				// ensure all tags are documented
				for _, tag := range op.Value.Tags {
					if _, ok := documentedTags[tag]; !ok {
						logging.Trace("tag is not documented, adding to document", "path", strings.ToUpper(op.Key)+" "+path.Key, "tag", tag)
						doc.Model.Tags = append(doc.Model.Tags, &base.Tag{Name: tag})
						documentedTags[tag] = true
					}
				}
			}
		}
	}

	return nil
}

var FixMissingSchemaTitlePatch = BuiltInPatcher{
	Type:                "builtin",
	ID:                  "fix-missing-schema-title",
	Description:         "Adds a title to all schemas that are missing a title",
	PatchV3DocumentFunc: FixMissingSchemaTitle,
}

// FixMissingSchemaTitle fills in missing schema titles with the schema key
func FixMissingSchemaTitle(doc *libopenapi.DocumentModel[v3.Document], config map[string]interface{}) error {
	// component schemas
	for schema := doc.Model.Components.Schemas.Oldest(); schema != nil; schema = schema.Next() {
		if schema.Value.Schema().Title == "" {
			schema.Value.Schema().Title = schema.Key
			logging.Trace("missing schema title, setting to schema key", "schema", schema.Key)
		}
	}

	// request bodies
	for rb := doc.Model.Components.RequestBodies.Oldest(); rb != nil; rb = rb.Next() {
		rbValue := rb.Value
		if rbValue == nil || rb.Value.Content == nil {
			continue
		}

		for mt := rbValue.Content.Oldest(); mt != nil; mt = mt.Next() {
			schemaRef := mt.Value.Schema
			if schemaRef != nil && schemaRef.Schema().Title == "" {
				schemaRef.Schema().Title = rb.Key
				logging.Trace("missing schema title in requestBody, setting to requestBody key", "requestBody", rb.Key, "mediaType", mt.Key)
			}
		}
	}

	// responses
	for resp := doc.Model.Components.Responses.Oldest(); resp != nil; resp = resp.Next() {
		respValue := resp.Value
		if respValue == nil || respValue.Content == nil {
			continue
		}

		for mt := respValue.Content.Oldest(); mt != nil; mt = mt.Next() {
			schemaRef := mt.Value.Schema
			if schemaRef != nil && schemaRef.Schema().Title == "" {
				schemaRef.Schema().Title = resp.Key
				logging.Trace("missing schema title in response, setting to response key", "response", resp.Key, "mediaType", mt.Key)
			}
		}
	}

	return nil
}

var FixRemoveCommonOperationIdPrefixPatch = BuiltInPatcher{
	Type:                "builtin",
	ID:                  "fix-remove-common-operation-id-prefix",
	Description:         "Removes common prefixes from operation IDs",
	PatchV3DocumentFunc: FixRemoveCommonOperationIdPrefix,
}

// FixRemoveCommonOperationIdPrefix sets the operation IDs of all operations and fixes some commonly seen issues.
func FixRemoveCommonOperationIdPrefix(doc *libopenapi.DocumentModel[v3.Document], config map[string]interface{}) error {
	var operationIds []string

	// scan all current operation IDs
	for path := doc.Model.Paths.PathItems.Oldest(); path != nil; path = path.Next() {
		for op := path.Value.GetOperations().Oldest(); op != nil; op = op.Next() {
			operationIds = append(operationIds, op.Value.OperationId)
		}
	}

	// detect common prefix
	commonPrefix := util.FindCommonStrPrefix(operationIds)
	if commonPrefix != "" {
		logging.Trace("removing common prefix from operation IDs", "prefix", commonPrefix)
		for path := doc.Model.Paths.PathItems.Oldest(); path != nil; path = path.Next() {
			for op := path.Value.GetOperations().Oldest(); op != nil; op = op.Next() {
				op.Value.OperationId = strings.TrimPrefix(op.Value.OperationId, commonPrefix)
			}
		}
	}

	return nil
}

var FixMissingOneOfFromDiscriminatorPatch = BuiltInPatcher{
	Type:                "builtin",
	ID:                  "fix-missing-oneof-from-discriminator",
	Description:         "Recursively populates missing oneOf lists using discriminator mapping entries across the entire document",
	PatchV3DocumentFunc: FixMissingOneOfFromDiscriminator,
}

func FixMissingOneOfFromDiscriminator(doc *libopenapi.DocumentModel[v3.Document], config map[string]interface{}) error {
	for schemaRef, schema := range openapidocument.CollectSchemas(doc) {
		if schema == nil || schema.Discriminator == nil {
			continue
		}

		// only fix if oneOf AND anyOf are missing.
		if len(schema.OneOf) > 0 || len(schema.AnyOf) > 0 {
			continue
		}

		// if discriminator is present without oneOf and without mapping, we cannot infer variants
		mapping := schema.Discriminator.Mapping
		if mapping == nil || mapping.Len() == 0 {
			slog.Warn("Discriminator present without composition and without mapping; cannot infer variants")
			continue
		}

		// if there is only one mapping entry, and it points to the same schema, skip to prevent circular oneOf
		if mapping.Len() == 1 {
			entry := mapping.Oldest()
			val := entry.Value
			if !strings.HasPrefix(val, "#") {
				val = "#/components/schemas/" + val
			}

			if val == schemaRef {
				slog.Debug("Skipping self-only discriminator mapping to prevent circular oneOf", "schema", schemaRef)
				schema.Discriminator = nil
				continue
			}
		}

		// track unique refs
		seenRefs := make(map[string]bool)
		if schema.OneOf == nil {
			schema.OneOf = make([]*base.SchemaProxy, 0)
		}
		for entry := mapping.Oldest(); entry != nil; entry = entry.Next() {
			refPath := entry.Value
			if refPath == "" || seenRefs[refPath] {
				continue
			}

			// normalize local references
			if !strings.HasPrefix(refPath, "#") {
				refPath = "#/components/schemas/" + refPath
				entry.Value = refPath
			}

			slog.Info("Synthesizing missing oneOf entry", "discriminatorValue", entry.Key, "targetRef", refPath)
			schema.OneOf = append(schema.OneOf, base.CreateSchemaProxyRef(refPath))
			seenRefs[refPath] = true
		}
	}

	return nil
}
