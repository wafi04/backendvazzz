package product

type ProductService struct {
	productRepo *ProductRepository
}

func NewProductService(repo *ProductRepository) *ProductService {
	return &ProductService{
		productRepo: repo,
	}
}

func (ser *ProductService) GetAll(categoryId int, subCategoryID int, role string) ([]ProductWithUserPrice, error) {
	return ser.productRepo.GetAll(categoryId, subCategoryID, role)
}
