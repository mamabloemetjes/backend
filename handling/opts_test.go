package handling

import (
	"net/http"
	"testing"
)

func TestParseProductListOptionsPreservesProductStatusFilters(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected string
	}{
		{name: "active", query: "?status=active", expected: "active"},
		{name: "inactive", query: "?status=inactive", expected: "inactive"},
		{name: "sold status alias", query: "?status=sold", expected: "sold"},
		{name: "sold with images", query: "?status=sold&include_images=true", expected: "sold"},
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
			if options.Status == nil {
				t.Fatal("expected status filter to be set")
			}
			if *options.Status != test.expected {
				t.Fatalf("expected status=%s, got %s", test.expected, *options.Status)
			}
		})
	}
}
