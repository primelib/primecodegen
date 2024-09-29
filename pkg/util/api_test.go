package util

import (
	"fmt"
	"testing"
)

func TestURLRemovePathParams(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"/users/{id}", "/users/"},
		{"/posts/{id}/comments/{commentId}", "/posts//comments/"},
		{"/{category}/{subcategory}", "//"},
		{"", ""},
		{"/path/without/params", "/path/without/params"},
	}

	for _, tc := range testCases {
		result := URLRemovePathParams(tc.input)
		if result != tc.expected {
			t.Errorf("URLRemovePathParams(%s) = %s; want %s", tc.input, result, tc.expected)
		}
	}
}

func TestURLPathParamAddByPrefix(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"/customers/{id}", "/customers/ById"},
		{"/orders/{orderNumber}", "/orders/ByOrderNumber"},
		{"/users/{userId}/details", "/users/ByUserId/details"},
		{"/products/{productId}/reviews", "/products/ByProductId/reviews"},
		{"/categories/{category}/items", "/categories/ByCategory/items"},
	}

	for _, test := range tests {
		result := URLPathParamAddByPrefix(test.input)
		if result != test.expected {
			t.Errorf("URLPathParamAddByPrefix(%q) = %q; want %q", test.input, result, test.expected)
		}
	}
}

func TestParseURLAPIVersion(t *testing.T) {
	testCases := []struct {
		url      string
		expected string
	}{
		{"/api/v1/books", "1"},
		{"/api/V2/books", "2"},
		{"/user/details", "1"},
		{"/order/v10/submit", "10"},
		{"", "1"},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("Input: %s", testCase.url), func(t *testing.T) {
			result := ParseURLAPIVersion(testCase.url)
			if result != testCase.expected {
				t.Errorf("Expected version: %s, but got: %s", testCase.expected, result)
			}
		})
	}
}

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
	}

	for _, tc := range testCases {
		result := ToOperationId(tc.method, tc.url)
		if result != tc.expectedID {
			t.Errorf("ToOperationId(%s, %s) = %s; want %s", tc.method, tc.url, result, tc.expectedID)
		}
	}
}
