# PrimeCodeGen

> Resolve specification issues and generate code from OpenAPI specifications.

This project is a collection of tools to help with merging, patching, and generating code (including user-provided templates) from API specifications.

- OpenAPI 3.0
- OpenAPI 3.1

## Installation

TODO: add installation instructions

## Commands

The following commands are available:

- `openapi-convert` - Convert OpenAPI specifications between different versions.
- `openapi-merge` - Combine multiple OpenAPI specifications into a single document.
- `openapi-patch` - Apply automatic modifications, merge multiple specifications, and incorporate custom patches.
- `openapi-export-template-data` - Extract and export template-related data useful for code generation from an OpenAPI specification.
- `openapi-generate` - Generate code from an OpenAPI specification.

### OpenAPI Convert

The `openapi-convert` command can be used to convert between different OpenAPI versions.

| Command                                                                                                   | Description                                                                             |
|-----------------------------------------------------------------------------------------------------------|-----------------------------------------------------------------------------------------|
| `primecodegen openapi-convert --format-in swagger20 --format-out openapi30 --input /in --output-dir /out` | Converts input - into output format (currently Swagger 2.0 to OpenAPI 3.0 is supported) |

**Note**: If `PRIMECODEGEN_SWAGGER_CONVERTER` is not set, the default swagger converter `https://converter.swagger.io/api/convert` will be used.

Environment Variables:

- `PRIMECODEGEN_SWAGGER_CONVERTER` - used to specify a custom [swagger-converter](https://github.com/swagger-api/swagger-converter) convert endpoint, used for Swagger 2.0 to OpenAPI 3.0 conversion.

### OpenAPI Merge

The `openapi-merge` command can be used to merge multiple OpenAPI specifications into one.

| Command                                                     | Description                                                                                                                                                                                                                               |
|-------------------------------------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `primecodegen openapi-merge --input /in --output-dir /out ` | Merge OpenAPI specifications to be compatible with code generation tool. Provide an empty OpenAPI 3.0 spec to build up a clean info-block. As an alternative use the built-in merge when using `openapi-patch` with multiple input specs. |

### OpenAPI Patch

The `openapi-patch` command can be used to apply automatic modifications, merge multiple specifications, and apply custom patches to the OpenAPI specification.

| Command                                                                            | Description                                                       |
|------------------------------------------------------------------------------------|-------------------------------------------------------------------|
| `primecodegen openapi-patch -i openapi.yaml -o patched.yaml`                       | if no patches are specified, the default ones are applied         |
| `primecodegen openapi-patch -i openapi.yaml -i openapi.part2.yaml -o patched.yaml` | merge one or more specifications into one                         |
| `primecodegen openapi-patch -i openapi.yaml -p flattenSchemas -o patched.yaml`     | apply built-in patch with id `flattenSchemas`                     |
| `primecodegen openapi-patch -i openapi.yaml -p json-patch:noservers.jsonpatch`     | apply a [jsonpatch](https://jsonpatch.com/) to the specification  |
| `primecodegen openapi-patch -i openapi.yaml -p git-patch:mypatch.patch`            | apply a `git patch` to the specification                          |
| `primecodegen openapi-patch -i openapi.yaml -p openapi-overlay:overlay.yaml`       | apply a openapi overlay                                           |
| `primecodegen openapi-patch validate openapi-overlay:dir/overlay.yaml`             | validate patch files (json-patch, git-patch, openapi-overlay, ... |
| `primecodegen openapi-patch list`                                                  | list all available patches                                        |

**Note**: All the options can be combined, e.g. merging multiple specifications, user-provided patches and built-in patchers.

The following built-in patches are available:

| Patch                             | Default | Description                                                                                                                                                   |
|-----------------------------------|---------|---------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `pruneOperationTags`              | false   | Removes all tags from operations.                                                                                                                             |
| `pruneOperationTagsExceptFirst`   | false   | Removes all tags from operations except the first one.                                                                                                        |
| `pruneCommonOperationIdPrefix`    | false   | Removes common operation id prefixes (e. g. all operationIds start with `API_`)                                                                               |
| `generateOperationIds`            | true    | Generates operationIds for all operations based on the HTTP path and method, overwriting existing ones.                                                       |
| `flattenSchema`                   | false   | Flattens inline request bodies and response schemas into the components section of the document.                                                              |
| `missingSchemaTitle`              | true    | Adds a title to all schemas that are missing a title.                                                                                                         |
| `createOperationTagsFromDocTitle` | false   | Removes all tags and creates one new tag per API spec from the document title, setting it on each operation. This patch will be applied before merging specs. |
| `inlineAllOfHierarchies`          | false   | Inlines properties of allOf-referenced schemas and removes allOf-references in schemas                                                                        |

**Note**: The patches are applied in the order you specify them in. `createOperationTagsFromDocTitle` is an exception to that rule because it is always applied first before specs are possibly merged. If none are specified, the patches flagged as `default` are applied.

### OpenAPI Template Data

The `openapi-generate-template` command can be used to pre-process the openapi spec and pass the resulting template data to an external code generator.
The command supports the following options:

| Command                                                                                      | Description                            |
|----------------------------------------------------------------------------------------------|----------------------------------------|
| `primecodegen openapi-export-template-data -i openapi.yaml -g go -t client`                  | generate go template data, stdout      |
| `primecodegen openapi-export-template-data -i openapi.yaml -g go -t client -o template.yaml` | generate go template data, file output |

### OpenAPI Code Generator

The `openapi-generate` command can be used to generate code from an OpenAPI specification, using a built-in or custom template.

| Command                                                                 | Description                                                   |
|-------------------------------------------------------------------------|---------------------------------------------------------------|
| `primecodegen openapi-generate -i openapi.yaml -g go -t client -o /out` | run code generation with generator `go` and template `client` |

Environment Variables:

- `PRIMECODEGEN_DEBUG_SPEC` - if set, the final OpenAPI specification is written to stdout.
- `PRIMECODEGEN_DEBUG_TEMPLATEDATA` - if set, the template data passed to the code generator is written to stdout.
- `PRIMECODEGEN_TEMPLATE_DIR` - if set, takes priority when looking for template files - useful for customizing templates.

## App

The `app` component provides a complete solution to maintain up-to-date API specifications and client libraries. (`GitHub Application` / `GitLab Application` / ...)

### Usage

| Commands                    | Description                                                                                        |
|-----------------------------|----------------------------------------------------------------------------------------------------|
| `primecodegen app-generate` | Creates a PR with updates to the OpenAPI Spec and the generated code.                              |
| `primecodegen app-release`  | Checks if the latest commit in the main branch has a release, automatically creating a tag if not. |

### Project Configuration

Projects are configured using a `primelib.yaml` file in the root of the repository.

**Example - Java**

```yaml
# yaml-language-server: $schema=https://raw.githubusercontent.com/primelib/primecodegen/main/configschema/primelib-v1.json
TODO: add example ...
```

### App Configuration

| Environment Variable     | Description                                                              |
|--------------------------|--------------------------------------------------------------------------|
| `PRIMEAPP_FOOTER_HIDE`   | Set to true to disable the footer note in the merge request description. |
| `PRIMEAPP_FOOTER_CUSTOM` | Set to replace the footer with your custom text.                         |

### Platform Configuration

You are *required* to have the environment variables for one platform set.

#### GitHub App

Create a GitHub App and configure it with the following permissions:

- Repository contents: Read & write
- Pull requests: Read & write
- Commit statuses: Read & write
- Checks: Read & write
- Metadata: Read-only

Create a private key and store it in a file.

| Environment Variable          | Description                       |
|-------------------------------|-----------------------------------|
| `GITHUB_APP_ID`               | The ID of the GitHub App.         |
| `GITHUB_APP_PRIVATE_KEY_FILE` | The path to the private key file. |

#### GitLab User

Create a GitLab user and generate a personal access token with the following permissions: `api`

| Environment Variable  | Description                |
|-----------------------|----------------------------|
| `GITLAB_SERVER`       | The GitLab server URL.     |
| `GITLAB_ACCESS_TOKEN` | The personal access token. |

## Roadmap

- [ ] Add support for AsyncAPI (https://github.com/asyncapi/parser-go/tree/master)
- [ ] Add support for Protobuf (https://github.com/yoheimuta/go-protoparser)

## Credits

- OpenAPI Parser: [pb33f/libopenapi](https://github.com/pb33f/libopenapi)
- Patches - OpenAPI Overlay: [speakeasy/openapi-overlay](https://github.com/speakeasy-api/openapi-overlay)
- Patches - Git: [bluekeyes/go-gitdiff](https://github.com/bluekeyes/go-gitdiff)
- Patches - JSON: [evanphx/jsonpatch](https://github.com/evanphx/json-patch)

## License

Released under the [MIT license](./LICENSE).
