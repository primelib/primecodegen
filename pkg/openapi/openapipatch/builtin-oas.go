package openapipatch

import (
	"github.com/pb33f/libopenapi"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

var FixOAS300VersionPatch = BuiltInPatcher{
	Type:                "builtin",
	ID:                  "fix-oas-300-version",
	Description:         "Fixes specs authored in OpenAPI 3.0.0 format but mistakenly labeled as a different version, without converting schema content.",
	PatchV3DocumentFunc: FixOAS300Version,
}

func FixOAS300Version(doc *libopenapi.DocumentModel[v3.Document], config map[string]interface{}) error {
	doc.Model.Version = "3.0.0"
	return nil
}

var FixOAS301VersionPatch = BuiltInPatcher{
	Type:                "builtin",
	ID:                  "fix-oas-301-version",
	Description:         "Fixes specs authored in OpenAPI 3.0.1 format but mistakenly labeled as a different version, without converting schema content.",
	PatchV3DocumentFunc: FixOAS301Version,
}

func FixOAS301Version(doc *libopenapi.DocumentModel[v3.Document], config map[string]interface{}) error {
	doc.Model.Version = "3.0.1"
	return nil
}

var FixOAS302VersionPatch = BuiltInPatcher{
	Type:                "builtin",
	ID:                  "fix-oas-302-version",
	Description:         "Fixes specs authored in OpenAPI 3.0.2 format but mistakenly labeled as a different version, without converting schema content.",
	PatchV3DocumentFunc: FixOAS302Version,
}

func FixOAS302Version(doc *libopenapi.DocumentModel[v3.Document], config map[string]interface{}) error {
	doc.Model.Version = "3.0.2"
	return nil
}

var FixOAS303VersionPatch = BuiltInPatcher{
	Type:                "builtin",
	ID:                  "fix-oas-303-version",
	Description:         "Fixes specs authored in OpenAPI 3.0.3 format but mistakenly labeled as a different version, without converting schema content.",
	PatchV3DocumentFunc: FixOAS303Version,
}

func FixOAS303Version(doc *libopenapi.DocumentModel[v3.Document], config map[string]interface{}) error {
	doc.Model.Version = "3.0.3"
	return nil
}

var FixOAS304VersionPatch = BuiltInPatcher{
	Type:                "builtin",
	ID:                  "fix-oas-304-version",
	Description:         "Fixes specs authored in OpenAPI 3.0.4 format but mistakenly labeled as a different version, without converting schema content.",
	PatchV3DocumentFunc: FixOAS304Version,
}

func FixOAS304Version(doc *libopenapi.DocumentModel[v3.Document], config map[string]interface{}) error {
	doc.Model.Version = "3.0.4"
	return nil
}

var FixOAS310VersionPatch = BuiltInPatcher{
	Type:                "builtin",
	ID:                  "fix-oas-310-version",
	Description:         "Fixes specs authored in OpenAPI 3.1.0 format but mistakenly labeled as a different version, without converting schema content.",
	PatchV3DocumentFunc: FixOAS310Version,
}

func FixOAS310Version(doc *libopenapi.DocumentModel[v3.Document], config map[string]interface{}) error {
	doc.Model.Version = "3.1.0"
	return nil
}

var FixOAS311VersionPatch = BuiltInPatcher{
	Type:                "builtin",
	ID:                  "fix-oas-311-version",
	Description:         "Fixes specs authored in OpenAPI 3.1.1 format but mistakenly labeled as a different version, without converting schema content.",
	PatchV3DocumentFunc: FixOAS311Version,
}

func FixOAS311Version(doc *libopenapi.DocumentModel[v3.Document], config map[string]interface{}) error {
	doc.Model.Version = "3.1.1"
	return nil
}
