package handling

import (
	"net/http"
	"testing"
)

func TestParseProductListOptionsPreservesBooleanStatusFilters(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected bool
	}{
		{name: "active", query: "?is_active=true", expected: true},
		{name: "sold", query: "?is_active=false", expected: false},
		{name: "sold status alias", query: "?status=sold", expected: false},
		{name: "sold with images", query: "?is_active=false&include_images=true", expected: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request, err := http.NewRequest(http.MethodGet, "/admin/products"+test.query, nil)
			if err != nil {
				t.Fatalf("create request: %v", err)
			}

			options, err := ParseProductListOptions(request)
			if err != nil {
				t.Fatalf("parse options: %v", err)
			}
			if options.IsActive == nil {
				t.Fatal("expected is_active filter to be set")
			}
			if *options.IsActive != test.expected {
				t.Fatalf("expected is_active=%t, got %t", test.expected, *options.IsActive)
			}
		})
	}
}
