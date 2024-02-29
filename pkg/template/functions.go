package template

import (
	"html/template"

	"github.com/primelib/primecodegen/pkg/util"
)

var templateFunctions = template.FuncMap{
	"firstNonEmpty": func(values ...string) string {
		return util.FirstNonEmptyString(values...)
	},
	"toLower": func(input string) string {
		return util.LowerCaseFirstLetter(input)
	},
	"toUpper": func(input string) string {
		return util.UpperCaseFirstLetter(input)
	},
	"trimNonASCII": func(input string) string {
		return util.TrimNonASCII(input)
	},
	"pascalCase": func(input string) string {
		return util.ToPascalCase(input)
	},
	"snakeCase": func(input string) string {
		return util.ToSnakeCase(input)
	},
	"kebabCase": func(input string) string {
		return util.ToKebabCase(input)
	},
	"camelCase": func(input string) string {
		return util.ToCamelCase(input)
	},
	"commentSingleLine": func(comment string) string {
		return util.CommentSingleLine(comment)
	},
	"conditionalValue": func(condition bool, trueValue, falseValue interface{}) interface{} {
		return util.ConditionalValue(condition, trueValue, falseValue)
	},
}
