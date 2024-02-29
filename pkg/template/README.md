# Template Information

## Type-Support by Iterator-Type

You need to add the following type-definition to the top of the template file for proper type-support. The type varies based on the iterator-type.

### Support File

```go
{{- /*gotype: github.com/primelib/primecodegen/pkg/openapi/openapigenerator.SupportOnceTemplate*/ -}}
```

### Type: ONCE_OPERATION

> Render once with all operations

```go

```

### Type: EACH_OPERATION

> Render each operation individually

```go
{{- /*gotype: github.com/primelib/primecodegen/pkg/openapi/openapigenerator.OperationEachTemplate*/ -}}
```

### Type: ONCE_MODEL

> Render once with all models

```go

```

### Type: EACH_MODEL

> Render each model individually

```go
{{- /*gotype: github.com/primelib/primecodegen/pkg/openapi/openapigenerator.ModelEachTemplate*/ -}}
```
