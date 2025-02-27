package util

import (
	"testing"
)

func TestToOperationId(t *testing.T) {
	testCases := []struct {
		method     string
		url        string
		expectedID string
	}{
		{"GET", "/api/users", "getUsersV1"},
		{"GET", "/api/users/{id}", "getUserByIdV1"},
		{"POST", "/api/posts/{id}/comments", "postPostByIdCommentsV1"},
		{"PUT", "/api/v1/updates/{id}", "putUpdateByIdV1"},
		{"DELETE", "/api/v2/posts/{id}", "deletePostByIdV2"},
		{"PATCH", "/api/v1/items/{id}/update", "patchItemByIdUpdateV1"},
		{"GET", "/api/v1/books/{bookId}/file/*", "getBookByBookIdFileV1"},
		{"PUT", "/api/v2/series/{seriesId}/read-progress/tachiyomi", "putSeriesBySeriesIdReadProgressTachiyomiV2"},
		{"PUT", "/api/v1.2/updates/{id}", "putV12UpdateByIdV1"},
	}

	for _, tc := range testCases {
		result := ToOperationId(tc.method, tc.url)
		if result != tc.expectedID {
			t.Errorf("ToOperationId(%s, %s) = %s; want %s", tc.method, tc.url, result, tc.expectedID)
		}
	}
}
