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

| Patch                           | Default | Description                                                                                             |
|---------------------------------|---------|---------------------------------------------------------------------------------------------------------|
| `pruneOperationTags`            | true    | Removes all tags from operations.                                                                       |
| `pruneOperationTagsExceptFirst` | false   | Removes all tags from operations except the first one.                                                  |
| `pruneCommonOperationIdPrefix`  | false   | Removes common operation id prefixes (e. g. all operationIds start with `API_`)                         |
| `generateOperationIds`          | false   | Generates operationIds for all operations based on the HTTP path and method, overwriting existing ones. |
| `flattenSchema`                 | true    | Flattens inline request bodies and response schemas into the components section of the document.        |
| `missingSchemaTitle`            | true    | Adds a title to all schemas that are missing a title.                                                   |

> Note: The patches are applied in the order you specify them in. If none are specified, patched flagged as `Default` are applied.

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
