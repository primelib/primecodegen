package openapipatch

import (
	"fmt"
	"testing"
)

func TestFindPrefix(t *testing.T) {
	testCases := []struct {
		input    []string
		expected string
	}{
		{[]string{"apple", "app", "ape"}, "ap"},
		{[]string{"apple", "banana", "peach"}, ""},
		{[]string{"hello", "hey", "hi"}, "h"},
		{[]string{"abcd", "abcde", "abcdef"}, "abcd"},
		{[]string{"same", "same", "same"}, "same"},
		{[]string{"", "hello", "hey"}, ""},
		{[]string{"prefix", "prefixes", "prefixed"}, "prefix"},
		{[]string{"/api/v1/books", "/api/v1/chapters", "/api/v1/authors"}, "/api/v1/"},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("Input: %v", testCase.input), func(t *testing.T) {
			result := findPrefix(testCase.input)
			if result != testCase.expected {
				t.Errorf("Expected prefix: %s, but got: %s", testCase.expected, result)
			}
		})
	}
}

func TestToOperationId(t *testing.T) {
	testCases := []struct {
		method   string
		url      string
		expected string
	}{
		{"GET", "/api/v1/books", "getBooksV1"},
		{"POST", "/api/v2/books", "postBooksV2"},
		{"PUT", "/v1/volumes/{volumeID}", "putVolumesV1"},
		{"DELETE", "/api/v2/users/authentication-activity", "deleteUsersAuthenticationActivityV2"},
		{"PATCH", "/api/v1/oauth2/providers", "patchOAuth2ProvidersV1"},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("Input: %s %s", testCase.method, testCase.url), func(t *testing.T) {
			result := toOperationId(testCase.method, testCase.url)
			if result != testCase.expected {
				t.Errorf("Expected operation ID: %s, but got: %s", testCase.expected, result)
			}
		})
	}
}

func TestExtractApiVersionVersionFromUrl(t *testing.T) {
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
			result := extractApiVersionVersionFromUrl(testCase.url)
			if result != testCase.expected {
				t.Errorf("Expected version: %s, but got: %s", testCase.expected, result)
			}
		})
	}
}
