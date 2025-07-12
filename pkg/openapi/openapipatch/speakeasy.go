package openapipatch

var SpeakeasyRemoveUnusedPatch = BuiltInPatcher{
	Type:        "speakeasy",
	ID:          "remove-unused",
	Description: "Given an OpenAPI file, remove all unused options",
}

var SpeakeasyCleanupPatch = BuiltInPatcher{
	Type:        "speakeasy",
	ID:          "cleanup",
	Description: "Cleanup the formatting of a given OpenAPI document",
}

var SpeakeasyFormatPatch = BuiltInPatcher{
	Type:        "speakeasy",
	ID:          "format",
	Description: "Format an OpenAPI document to be more human-readable",
}

var SpeakeasyNormalizePatch = BuiltInPatcher{
	Type:        "speakeasy",
	ID:          "normalize",
	Description: "Normalize an OpenAPI document to be more human-readable",
}
