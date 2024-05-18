package commonpatch

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestApplyGitPatch(t *testing.T) {
	input := []byte(`MIT License

Copyright (c) 0000 John Doe

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:
...`)
	patchContent := []byte(`diff --git a/LICENSE.md b/LICENSE.md
index d8ca62b..183131e 100644
--- a/LICENSE.md
+++ b/LICENSE.md
@@ -1,6 +1,6 @@
 MIT License
 
-Copyright (c) 0000 John Doe
+Copyright (c) 0000 Jane Doe
 
 Permission is hereby granted, free of charge, to any person obtaining a copy
 of this software and associated documentation files (the "Software"), to deal
`)
	expected := `MIT License

Copyright (c) 0000 Jane Doe

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:
...`

	result, err := ApplyGitPatch(input, patchContent)
	assert.NoError(t, err)
	assert.Equal(t, expected, string(result))
}
