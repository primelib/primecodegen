package openapi_java

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCleanJavaImports(t *testing.T) {
	content := []byte(`
		import java.util.List;
		import java.util.Map;
		import java.util.ArrayList;
		
		public class Test {
			public static void main(String[] args) {
				List<String> list = new ArrayList<>();
			}
		}
	`)
	expected := []byte(`
		import java.util.List;
		import java.util.ArrayList;
		
		public class Test {
			public static void main(String[] args) {
				List<String> list = new ArrayList<>();
			}
		}
	`)
	result := CleanJavaImports(content)
	assert.Equal(t, string(expected), string(result))
}

func TestFindImports(t *testing.T) {
	result := findImports([][]byte{
		[]byte("import java.util.List;"),
		[]byte("import java.util.Map;"),
		[]byte("import java.util.Set;"),
		[]byte("other line"),
	})

	assert.Len(t, result, 3)
	assert.Equal(t, "java.util.List", result[0])
	assert.Equal(t, "java.util.Map", result[1])
	assert.Equal(t, "java.util.Set", result[2])
}

func TestFindUnusedImports(t *testing.T) {
	content := []byte(`
		import java.util.List;
		import java.util.Map;
		import java.util.ArrayList;
		
		public class Test {
			public static void main(String[] args) {
				List<String> list = new ArrayList<>();
			}
		}
	`)
	imports := []string{
		"java.util.List",
		"java.util.Map",
		"java.util.ArrayList",
	}

	result := findUnusedImports(content, imports)
	assert.Len(t, result, 1)
	assert.Equal(t, "java.util.Map", result[0])
}
