package model

type SubCategory struct {
	Id         int    `json:"id" db:"id"`
	CategoryId int    `json:"categoryId" db:"category_id"`
	Code       string `json:"code" db:"code"`
	Name       string `json:"name" db:"name"`
	Status     string `json:"status" db:"status"`
}

type CreateSubcategory struct {
	CategoryId int    `json:"categoryId" validate:"required"`
	Code       string `json:"code" validate:"required"`
	Name       string `json:"name" validate:"required"`
	Status     string `json:"status" validate:"required,oneof=active inactive"`
}

type UpdateSubcategory struct {
	CategoryId *int    `json:"categoryId,omitempty"`
	Code       *string `json:"code,omitempty"`
	Name       *string `json:"name,omitempty"`
	Status     *string `json:"status,omitempty" validate:"omitempty,oneof=active inactive"`
}
