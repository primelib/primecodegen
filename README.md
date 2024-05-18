# PrimeCodeGen

> Resolve specification issues and generate code from OpenAPI specifications.

This project is a collection of tools to help with merging, patching, and generating code (including user-provided templates) from API specifications.

- OpenAPI 3.0
- OpenAPI 3.1

## Installation

TODO: add installation instructions

## OpenAPI Generate

The `openapi-generate` command can be used to generate code from an OpenAPI specification, using a built-in or custom template.
You can also use the `openapi-generate-template` command to generate template data for custom external code generators.
The command supports the following options:

| Command                                                                                      | Description                                                      |
|----------------------------------------------------------------------------------------------|------------------------------------------------------------------|
| `primecodegen openapi-export-template-data -i openapi.yaml -g go -t client`                  | generate go template data, stdout                                |
| `primecodegen openapi-export-template-data -i openapi.yaml -g go -t client -o template.yaml` | generate go template data, file output                           |
| `primecodegen openapi-generate -i openapi.yaml -g go -t client -o /out`                      | run code generation with generator `go` and template `client`    |

## OpenAPI Patch

The `openapi-patch` command can be used to apply automatic modifications, merge multiple specifications, and apply custom patches to the OpenAPI specification.

| Command                                                                                      | Description                                                      |
|----------------------------------------------------------------------------------------------|------------------------------------------------------------------|
| `primecodegen openapi-patch -i openapi.yaml -o patched.yaml`                                 | if no patches are specified, the default ones are applied        |
| `primecodegen openapi-patch -i openapi.yaml -i openapi.part2.yaml -o patched.yaml`           | merge one or more specifications into one                        |
| `primecodegen openapi-patch -i openapi.yaml -p flattenSchemas -o patched.yaml`               | apply built-in patch with id `flattenSchemas`                    |
| `primecodegen openapi-patch -i openapi.yaml -f noservers.jsonpatch`                          | apply a [jsonpatch](https://jsonpatch.com/) to the specification |
| `primecodegen openapi-patch -i openapi.yaml -f mypatch.patch`                                | apply a `git patch` to the specification                         |
| `primecodegen openapi-patch list`                                                            | list available patches                                           |

**Note**: All the options can be combined, e.g. merging multiple specifications, custom user-provided patches and built-in patches.

The following built-in patches are available:

| Patch                           | Default | Description                                                                                             |
|---------------------------------|---------|---------------------------------------------------------------------------------------------------------|
| `pruneOperationTags`            | true    | Removes all tags from operations.                                                                       |
| `pruneOperationTagsExceptFirst` | false   | Removes all tags from operations except the first one.                                                  |
| `pruneCommonOperationIdPrefix`  | false   | Removes common operation id prefixes (e. g. all operationIds start with `API_`)                         |
| `generateOperationIds`          | false   | Generates operationIds for all operations based on the HTTP path and method, overwriting existing ones. |
| `flattenSchema`                 | true    | Flattens inline request bodies and response schemas into the components section of the document.        |
| `missingSchemaTitle`            | true    | Adds a title to all schemas that are missing a title.                                                   |

> Note: The patches are applied in the order you specify them in. If none are specified, the patches flagged as `default` are applied.

## OpenAPI Template Data

The `openapi-generate-template` command can be used to pre-process the openapi spec and pass the resulting template data to an external code generator.

TODO: documentation

## OpenAPI Code Generator

The `openapi-generate` command runs the code generation process.

TODO: documentation

## Roadmap

- [ ] Add support for AsyncAPI (https://github.com/asyncapi/parser-go/tree/master)
- [ ] Add support for Protobuf (https://github.com/yoheimuta/go-protoparser)

## Credits

- OpenAPI Parser: [libopenapi](https://github.com/pb33f/libopenapi)
- Patches - Git: [go-gitdiff](https://github.com/bluekeyes/go-gitdiff)
- Patches - JSON: [jsonpatch](https://github.com/evanphx/json-patch)

## License

Released under the [MIT license](./LICENSE).
