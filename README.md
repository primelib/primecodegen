# PrimeCodeGen

> Resolve specification issues and generate code from OpenAPI specifications.

Supported specifications:

- OpenAPI 3.0
- OpenAPI 3.1

## Installation

TODO: add installation instructions

## Command Line Interface

The CLI supports three use-cases:

- patch common issues / normalize the openapi spec pre code generation
- generate template data for custom external code generators
- generate code using built in templates

| Command                                                                                      | Description                                                   |
|----------------------------------------------------------------------------------------------|---------------------------------------------------------------|
| `primecodegen openapi-patch -i openapi.yaml -o patched.yaml`                                 | apply automatic modifications and fixes to the openapi spec   |
| `primecodegen openapi-patch -i openapi.yaml -p flattenSchemas -o patched.yaml`               | apply patch with id `flattenSchemas`                          |
| `primecodegen openapi-patch -l`                                                              | list available patches                                        |
| `primecodegen openapi-export-template-data -i openapi.yaml -g go -t client`                  | generate go template data, stdout                             |
| `primecodegen openapi-export-template-data -i openapi.yaml -g go -t client -o template.yaml` | generate go template data, file output                        |
| `primecodegen openapi-generate -i openapi.yaml -g go -t client -o /out`                      | run code generation with generator `go` and template `client` |

## OpenAPI Patch

The `openapi-patch` command applies modifications and fixes to the openapi spec.

**pruneOperationTags**

Removes all tags from operations.

**pruneCommonOperationIdPrefix**

Removes common prefixes from operation IDs. If you have a spec where all operationIds start with e.g. `API_`, this patch will remove that prefix from all operationIds.
If you have a lot of bad operationIds, using `generateOperationIds` might be a better option.

**generateOperationIds**

Generates operationIds for all operations and overwrites existing ones. The operationId is generated based on the HTTP method and the path.

**flattenSchema**

Flattens inline request bodies and response schemas into the components section of the document.

**missingSchemaTitle**

Adds a title to all schemas that are missing a title.

> Note: The patches are applied in the order you specify them in.

## OpenAPI Template Data

The `openapi-generate-template` command can be used to pre-process the openapi spec and pass the resulting template data to an external code generator.

TODO: documentation

## OpenAPI Code Generator

The `openapi-generate` command runs the code generation process.

TODO: documentation

## Roadmap

- [ ] Add support for AsyncAPI

## Credits

- OpenAPI Parser: [libopenapi](https://github.com/pb33f/libopenapi)

## License

Released under the [MIT license](./LICENSE).
