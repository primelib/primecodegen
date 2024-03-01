package openapidocument

import (
	"fmt"

	"github.com/pb33f/libopenapi/datamodel/high/base"
)

func IsMutable(schema *base.Schema) (bool, error) {
	// see https://azure.github.io/autorest/extensions/#x-ms-mutability
	if mxMutability, ok := schema.Extensions.Get("x-ms-mutability"); ok {
		var values []string
		err := mxMutability.Decode(&values)
		if err != nil {
			return false, fmt.Errorf("unable to decode x-ms-mutability: %w", err)
		}

		// possible values: create, read, update

	}

	return false, nil
}
