package catalog

import (
	"context"
	"regexp"
	"strconv"
	"strings"
)

var (
	slugPattern     = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
	skuPattern      = regexp.MustCompile(`^[A-Z0-9][A-Z0-9._-]*$`)
	currencyPattern = regexp.MustCompile(`^[A-Z]{3}$`)
)

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) CreateProduct(ctx context.Context, input CreateProductInput) (Product, error) {
	normalizeInput(&input)
	if validationErr := validateCreateInput(input); validationErr != nil {
		return Product{}, validationErr
	}
	return s.repository.Create(ctx, input)
}

func (s *Service) ListProducts(ctx context.Context, page, pageSize int32) (Page, error) {
	fields := map[string]string{}
	if page < 1 {
		fields["page"] = "must be at least 1"
	}
	if pageSize < 1 || pageSize > 100 {
		fields["page_size"] = "must be between 1 and 100"
	}
	if len(fields) > 0 {
		return Page{}, &ValidationError{Fields: fields}
	}
	return s.repository.List(ctx, page, pageSize)
}

func (s *Service) GetProduct(ctx context.Context, slug string) (Product, error) {
	slug = strings.TrimSpace(slug)
	if len(slug) < 1 || len(slug) > 200 || !slugPattern.MatchString(slug) {
		return Product{}, ErrNotFound
	}
	return s.repository.GetBySlug(ctx, slug)
}

func normalizeInput(input *CreateProductInput) {
	input.Name = strings.TrimSpace(input.Name)
	input.Slug = strings.TrimSpace(input.Slug)
	input.Description = strings.TrimSpace(input.Description)
	for index := range input.SKUs {
		input.SKUs[index].Code = strings.TrimSpace(input.SKUs[index].Code)
		input.SKUs[index].Currency = strings.TrimSpace(input.SKUs[index].Currency)
	}
}

func validateCreateInput(input CreateProductInput) error {
	fields := map[string]string{}
	if len(input.Name) < 1 || len(input.Name) > 200 {
		fields["name"] = "must contain between 1 and 200 characters"
	}
	if len(input.Slug) < 1 || len(input.Slug) > 200 || !slugPattern.MatchString(input.Slug) {
		fields["slug"] = "must be lowercase words separated by single hyphens"
	}
	if len(input.Description) > 5000 {
		fields["description"] = "must not exceed 5000 characters"
	}
	if len(input.SKUs) < 1 || len(input.SKUs) > 100 {
		fields["skus"] = "must contain between 1 and 100 items"
	}
	seenCodes := make(map[string]struct{}, len(input.SKUs))
	for index, sku := range input.SKUs {
		prefix := "skus[" + strconv.Itoa(index) + "]"
		if len(sku.Code) < 1 || len(sku.Code) > 100 || !skuPattern.MatchString(sku.Code) {
			fields[prefix+".code"] = "must contain only uppercase letters, digits, dots, underscores, or hyphens"
		}
		if _, duplicate := seenCodes[sku.Code]; duplicate {
			fields[prefix+".code"] = "must be unique within the product"
		}
		seenCodes[sku.Code] = struct{}{}
		if sku.PriceCents < 0 {
			fields[prefix+".price_cents"] = "must be non-negative"
		}
		if !currencyPattern.MatchString(sku.Currency) {
			fields[prefix+".currency"] = "must be a three-letter uppercase ISO 4217 code"
		}
	}
	if len(fields) > 0 {
		return &ValidationError{Fields: fields}
	}
	return nil
}
