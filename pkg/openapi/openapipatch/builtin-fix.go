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
						slog.Debug("moving property from additionalProperties to properties", "schema", schema.Key, "property", p.Key)
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
					slog.Debug("fixing maximum value for integer", "schema", schema.Key, "property", p.Key)
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
				slog.Debug("operation is missing tags, adding default tag", "path", strings.ToUpper(op.Key)+" "+path.Key)
				op.Value.Tags = append(op.Value.Tags, "default")
			} else {
				// ensure all tags are documented
				for _, tag := range op.Value.Tags {
					if _, ok := documentedTags[tag]; !ok {
						slog.Debug("tag is not documented, adding to document", "path", strings.ToUpper(op.Key)+" "+path.Key, "tag", tag)
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
			slog.Debug("missing schema title, setting to schema key", "schema", schema.Key)
		}
	}

	// request bodies
	for rb := doc.Model.Components.RequestBodies.Oldest(); rb != nil; rb = rb.Next() {
		rbValue := rb.Value
		if rbValue == nil {
			continue
		}

		for mt := rbValue.Content.Oldest(); mt != nil; mt = mt.Next() {
			schemaRef := mt.Value.Schema
			if schemaRef != nil && schemaRef.Schema().Title == "" {
				schemaRef.Schema().Title = rb.Key
				slog.Debug("missing schema title in requestBody, setting to requestBody key", "requestBody", rb.Key, "mediaType", mt.Key)
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
		slog.Debug("removing common prefix from operation IDs", "prefix", commonPrefix)
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
	Description:         "Adds missing oneOf entries based on discriminator.mapping when discriminator is present but oneOf is missing.",
	PatchV3DocumentFunc: FixMissingOneOfFromDiscriminator,
}

func FixMissingOneOfFromDiscriminator(doc *libopenapi.DocumentModel[v3.Document], config map[string]interface{}) error {
	walkedDoc := openapidocument.WalkDocument(doc)
	if len(walkedDoc.Schemas) == 0 {
		return nil
	}

	for _, schema := range openapidocument.CollectSchemas(doc) {
		if schema == nil {
			continue
		}

		// check if discriminator is missing or oneOf is already present
		if schema.Discriminator == nil || len(schema.OneOf) > 0 {
			continue
		}

		// discriminator without mapping is too ambiguous to recover
		if schema.Discriminator.Mapping == nil || schema.Discriminator.Mapping.Len() == 0 {
			slog.Warn("discriminator present without oneOf and without mapping; cannot infer variants")
			continue
		}

		// Build oneOf from mapping values
		for entry := schema.Discriminator.Mapping.Oldest(); entry != nil; entry = entry.Next() {
			if entry.Value == "" {
				continue
			}

			slog.With("discriminatorKey", entry.Key).With("schemaRef", entry.Value).Info("adding missing oneOf entry from discriminator mapping")
			if schema.OneOf == nil {
				schema.OneOf = make([]*base.SchemaProxy, 0)
			}
			schema.OneOf = append(schema.OneOf, base.CreateSchemaProxyRef(entry.Value))
			slog.With("current oneOf", schema.OneOf).Warn("added oneOf entry to discriminator mapping")
		}
	}

	return nil
}
