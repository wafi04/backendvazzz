package subcategory

import (
	"context"

	"github.com/wafi04/backendvazzz/pkg/model"
)

// Service Layer
type SubCategoryService struct {
	subCategoryRepo *SubCategoryRepository
}

func NewSubCategoryService(subCategoryRepo *SubCategoryRepository) *SubCategoryService {
	return &SubCategoryService{
		subCategoryRepo: subCategoryRepo,
	}
}

// Create SubCategory
func (s *SubCategoryService) CreateSubCategory(ctx context.Context, data model.CreateSubcategory) (*model.SubCategory, error) {
	// Check if category exists

	return s.subCategoryRepo.Create(ctx, data)
}

// Get All SubCategories
func (s *SubCategoryService) GetAllSubCategories(ctx context.Context, skip, limit int, search, status string) ([]model.SubCategory, int, error) {
	return s.subCategoryRepo.GetAll(ctx, skip, limit, search, status)
}

// Get SubCategories by Category ID
func (s *SubCategoryService) GetSubCategoriesByCategoryID(ctx context.Context, categoryID, skip, limit int, search, status string) ([]model.SubCategory, int, error) {
	return s.subCategoryRepo.GetByCategoryID(ctx, categoryID, skip, limit, search, status)
}

// Get model.SubCategory by ID
func (s *SubCategoryService) GetSubCategoryByID(ctx context.Context, id int) (*model.SubCategory, error) {
	return s.subCategoryRepo.GetByID(ctx, id)
}

// Update model.SubCategory
func (s *SubCategoryService) UpdateSubCategory(ctx context.Context, id int, data model.UpdateSubcategory) (*model.SubCategory, error) {
	// Check if subcategory exists
	_, err := s.subCategoryRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.subCategoryRepo.Update(ctx, id, data)
}

// Delete SubCategory (Soft Delete)
func (s *SubCategoryService) DeleteSubCategory(ctx context.Context, id int) error {
	return s.subCategoryRepo.Delete(ctx, id)
}
