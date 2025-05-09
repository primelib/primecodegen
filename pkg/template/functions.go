package template

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"text/template"

	"github.com/primelib/primecodegen/pkg/util"
	"gopkg.in/yaml.v3"
)

var templateFunctions = template.FuncMap{
	"hasPrefix": func(s, prefix string) bool {
		return strings.HasPrefix(s, prefix)
	},
	"hasSuffix": func(s, suffix string) bool {
		return strings.HasSuffix(s, suffix)
	},
	"firstNonEmpty": func(values ...string) string {
		return util.FirstNonEmptyString(values...)
	},
	"lowerCase": func(input string) string {
		return strings.ToLower(input)
	},
	"upperCase": func(input string) string {
		return strings.ToUpper(input)
	},
	"lowerCaseFirstLetter": func(input string) string {
		return util.LowerCaseFirstLetter(input)
	},
	"upperCaseFirstLetter": func(input string) string {
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
	"slug": func(input string) string {
		return util.ToSlug(input)
	},
	"commentSingleLine": func(input string) string {
		return util.CommentSingleLine(input)
	},
	"commentMultiLine": func(prefix, input string) string {
		return util.CommentMultiLine(input, prefix)
	},
	"escapeJavadoc": func(input string) string {
		input = strings.Replace(input, "<", "&lt;", -1)
		input = strings.Replace(input, ">", "&gt;", -1)
		return input
	},
	"wrapIn": func(left string, right string, input string) string {
		return left + input + right
	},
	"conditionalValue": func(condition bool, trueValue, falseValue interface{}) interface{} {
		return util.Ternary(condition, trueValue, falseValue)
	},
	"marshalJSON": func(input interface{}) string {
		a, _ := json.Marshal(input)
		return string(a)
	},
	"marshalYAML": func(input interface{}) string {
		a, _ := yaml.Marshal(input)
		return string(a)
	},
	"isEmpty": func(input string) bool {
		return input == ""
	},
	"isNotEmpty": func(input string) bool {
		return input != ""
	},
	// toFilePath is used to convert a package path into a file path (e.g. "io.github.myuser" -> "io/github/myuser")
	"toFilePath": func(input string) string {
		return strings.ReplaceAll(input, ".", string(os.PathSeparator))
	},
	// notLast is used to determine if the current index is not the last index in a slice
	"notLast": func(data interface{}, idx interface{}) bool {
		val := reflect.ValueOf(data)
		if val.Kind() == reflect.Slice {
			idxInt, ok := idx.(int)
			if !ok {
				return false
			}

			return idxInt < val.Len()-1
		} else if val.Kind() == reflect.Map {
			idxStr, ok := idx.(string)
			if !ok {
				return false
			}

			return idxStr != val.MapKeys()[val.Len()-1].String()
		}

		return false
	},
}
