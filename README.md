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

| Command                                                                                   | Description                                                   |
|-------------------------------------------------------------------------------------------|---------------------------------------------------------------|
| `primecodegen openapi-patch -i openapi.yaml -o patched.yaml`                              | apply automatic modifications and fixes to the openapi spec   |
| `primecodegen openapi-generate-template -i openapi.yaml -g go -t client`                  | generate go template data, stdout                             |
| `primecodegen openapi-generate-template -i openapi.yaml -g go -t client -o template.yaml` | generate go template data, file output                        |
| `primecodegen openapi-generate -i openapi.yaml -g go -t client -o /out`                   | run code generation with generator `go` and template `client` |

## OpenAPI Patch

The `openapi-patch` command applies automatic modifications and fixes to the openapi spec.

- pruneOperationTags - Remove all tags from operations
- generateOperationIds - Generate operationIds for all operations and overwrite existing ones

## OpenAPI Generate Template

The `openapi-generate-template` command generates template data that can be used to build your own code generator templates without having to deal with most of the complexity of the openapi spec.

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
