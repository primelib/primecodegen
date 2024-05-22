package openapi_java

import (
	"bytes"
	"strings"
)

// CleanJavaImports removes unused imports from Java files
func CleanJavaImports(content []byte) []byte {
	lines := bytes.Split(content, []byte("\n"))

	// find all imports
	imports := findImports(lines)

	// find unused imports
	unusedImports := findUnusedImports(content, imports)

	// remove unused imports from content
	var output []byte
	for i, line := range lines {
		removeLine := false
		for _, unusedImport := range unusedImports {
			if bytes.Contains(line, []byte("import "+unusedImport+";")) {
				removeLine = true
				break
			}
		}

		if !removeLine {
			output = append(output, line...)
			if i < len(lines)-1 {
				output = append(output, '\n')
			}
		}
	}

	return output
}

// findImports returns a list of imports
func findImports(lines [][]byte) (imports []string) {
	// find all imports
	for _, line := range lines {
		line = bytes.TrimSpace(line)
		if bytes.HasPrefix(line, []byte("import ")) {
			importPath := bytes.TrimSpace(bytes.TrimSuffix(bytes.TrimPrefix(line, []byte("import ")), []byte(";")))
			imports = append(imports, string(importPath))

		}
	}

	return imports
}

// findUnusedImports returns a list of unused imports
func findUnusedImports(content []byte, imports []string) (unusedImports []string) {
	for _, importLine := range imports {
		className := getLastPart(importLine)
		if className == "*" {
			continue
		}

		patterns := []string{
			"new " + className,
			className + ".",
			className + " ",
			"@" + className,
			className + "(",
			className + "::",
			className + "<",
			"<" + className + ">",
		}

		isUsed := false
		for _, pattern := range patterns {
			if bytes.Contains(content, []byte(pattern)) {
				isUsed = true
				break
			}
		}

		if !isUsed {
			unusedImports = append(unusedImports, importLine)
		}
	}

	return unusedImports
}

func getLastPart(input string) string {
	if i := strings.LastIndex(input, "."); i != -1 {
		return input[i+1:]
	}
	return input
}
