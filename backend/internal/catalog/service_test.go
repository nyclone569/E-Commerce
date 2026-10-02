package catalog

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

type stubRepository struct {
	createdInput CreateProductInput
	createErr    error
	product      Product
	getSlug      string
	getErr       error
	page         Page
}

func (r *stubRepository) Create(_ context.Context, input CreateProductInput) (Product, error) {
	r.createdInput = input
	return Product{Name: input.Name, Slug: input.Slug}, r.createErr
}

func (r *stubRepository) List(_ context.Context, _, _ int32) (Page, error) {
	return r.page, nil
}

func (r *stubRepository) GetBySlug(_ context.Context, slug string) (Product, error) {
	r.getSlug = slug
	return r.product, r.getErr
}

func (r *stubRepository) GetSKUs(_ context.Context, _ []uuid.UUID) ([]SKUDetails, error) {
	return nil, nil
}

func TestCreateProductNormalizesAndDelegates(t *testing.T) {
	repository := &stubRepository{}
	service := NewService(repository)

	product, err := service.CreateProduct(context.Background(), CreateProductInput{
		Name: "  Aurora Mug  ", Slug: "aurora-mug", Description: "  Ceramic  ",
		SKUs: []CreateSKUInput{{Code: " MUG-WHITE ", PriceMinor: 129900, Currency: " VND "}},
	})
	if err != nil {
		t.Fatalf("CreateProduct() error = %v", err)
	}
	if product.Name != "Aurora Mug" || repository.createdInput.Description != "Ceramic" {
		t.Fatalf("input was not normalized: %#v", repository.createdInput)
	}
}

func TestCreateProductRejectsInvalidFields(t *testing.T) {
	service := NewService(&stubRepository{})
	_, err := service.CreateProduct(context.Background(), CreateProductInput{
		Name: "", Slug: "Not Valid", SKUs: []CreateSKUInput{{Code: "bad code", PriceMinor: -1, Currency: "usd"}},
	})
	var validationErr *ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("CreateProduct() error = %v, want ValidationError", err)
	}
	for _, field := range []string{"name", "slug", "skus[0].code", "skus[0].price_minor", "skus[0].currency"} {
		if _, exists := validationErr.Fields[field]; !exists {
			t.Errorf("missing validation error for %s", field)
		}
	}
}

func TestCreateProductRejectsDuplicateSKUCodes(t *testing.T) {
	service := NewService(&stubRepository{})
	_, err := service.CreateProduct(context.Background(), CreateProductInput{
		Name: "Mug", Slug: "mug", SKUs: []CreateSKUInput{
			{Code: "MUG", PriceMinor: 100000, Currency: "VND"},
			{Code: "MUG", PriceMinor: 200000, Currency: "VND"},
		},
	})
	var validationErr *ValidationError
	if !errors.As(err, &validationErr) || validationErr.Fields["skus[1].code"] == "" {
		t.Fatalf("CreateProduct() error = %v, want duplicate SKU error", err)
	}
}

func TestListProductsValidatesPagination(t *testing.T) {
	service := NewService(&stubRepository{})
	_, err := service.ListProducts(context.Background(), 0, 101)
	var validationErr *ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("ListProducts() error = %v, want ValidationError", err)
	}
	if validationErr.Fields["page"] == "" || validationErr.Fields["page_size"] == "" {
		t.Fatalf("unexpected fields: %#v", validationErr.Fields)
	}
}

func TestGetProductDelegatesValidSlug(t *testing.T) {
	repository := &stubRepository{product: Product{Name: "Aurora Mug", Slug: "aurora-mug"}}
	service := NewService(repository)

	product, err := service.GetProduct(context.Background(), "  aurora-mug  ")
	if err != nil {
		t.Fatalf("GetProduct() error = %v", err)
	}
	if repository.getSlug != "aurora-mug" || product.Slug != "aurora-mug" {
		t.Fatalf("unexpected lookup: slug=%q product=%#v", repository.getSlug, product)
	}
}

func TestGetProductHidesInvalidSlugAsNotFound(t *testing.T) {
	service := NewService(&stubRepository{})
	_, err := service.GetProduct(context.Background(), "Not A Slug")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("GetProduct() error = %v, want ErrNotFound", err)
	}
}
