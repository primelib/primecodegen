package util

import "testing"

func TestOpenAPIOperationSlug(t *testing.T) {
	testCases := []struct {
		name     string
		method   string
		path     string
		expected string
	}{
		{
			name:     "normal path",
			method:   "GET",
			path:     "/v1/myopname/create",
			expected: "get-v1-myopname-create",
		},
		{
			name:     "path params",
			method:   "POST",
			path:     "/v1/users/{userId}/bookings/{bookingId}",
			expected: "post-v1-users-userid-bookings-bookingid",
		},
		{
			name:     "mixed case and special chars",
			method:   "Patch",
			path:     "/V1/MyOpName/create-order",
			expected: "patch-v1-myopname-create-order",
		},
		{
			name:     "root path",
			method:   "DELETE",
			path:     "/",
			expected: "delete",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := OpenAPIOperationSlug(tc.method, tc.path)
			if actual != tc.expected {
				t.Errorf("OpenAPIOperationSlug(%q, %q) = %q; expected %q", tc.method, tc.path, actual, tc.expected)
			}
		})
	}
}
