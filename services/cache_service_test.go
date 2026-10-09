package services

import (
	"testing"

	"github.com/google/uuid"
)

func TestProductCacheKeysSeparateProductVariants(t *testing.T) {
	id := uuid.New()

	withImages := productIDCacheKey(id, true)
	withoutImages := productIDCacheKey(id, false)
	if withImages == withoutImages {
		t.Fatal("product ID cache keys must separate image variants")
	}

	allProducts := activeProductsCacheKey(1, 20, true, "")
	typedProducts := activeProductsCacheKey(1, 20, true, "bouquet")
	if allProducts == typedProducts {
		t.Fatal("active product cache keys must separate product types")
	}
	if allProducts != "products:active:page:1:size:20:type:all:images:true" {
		t.Fatalf("unexpected all-products cache key: %s", allProducts)
	}
}

func TestProductCacheKeyFormatsRemainStable(t *testing.T) {
	id := uuid.New()

	if got, want := productSKUCacheKey("SKU-1"), "product:sku:SKU-1"; got != want {
		t.Fatalf("SKU cache key = %q, want %q", got, want)
	}
	if got, want := productCountCacheKey("active"), "products:count:active"; got != want {
		t.Fatalf("count cache key = %q, want %q", got, want)
	}
	if got, want := productIDCachePattern(id), "product:id:"+id.String()+":*"; got != want {
		t.Fatalf("ID cache pattern = %q, want %q", got, want)
	}
}
