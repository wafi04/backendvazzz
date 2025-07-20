package types

type Category struct {
	ID              int    `json:"id"`
	Name            string `json:"name"`
	SubName         string `json:"sub_name"`
	Brand           string `json:"brand"`
	Code            string `json:"code"`
	IsCheckNickname string `json:"is_check_nickname"`
	Status          string `json:"status"`
	Thumbnail       string `json:"thumbnail"`
	Type            string `json:"type"`
	Banner          string `json:"banner"`
	Instruction     string `json:"instruction"`
	Information     string `json:"information"`
	Placeholder1    string `json:"placeholder_1"`
	Placeholder2    string `json:"placeholder_2"`
	CreatedAt       string `json:"createdAt"`
	UpdatedAt       string `json:"updatedAt"`
}

type CreateCategory struct {
	Name            string `json:"name"`
	SubName         string `json:"sub_name"`
	Brand           string `json:"brand"`
	Code            string `json:"code"`
	IsCheckNickname string `json:"is_check_nickname"`
	Status          string `json:"status"`
	Thumbnail       string `json:"thumbnail"`
	Type            string `json:"type"`
	Banner          string `json:"banner"`
	Placeholder1    string `json:"placeholder_1"`
	Placeholder2    string `json:"placeholder_2"`
	Instruction     string `json:"instruction"`
	Information     string `json:"information"`
}

type UpdateCategory struct {
	Name            *string `json:"name"`
	SubName         *string `json:"sub_name"`
	Brand           *string `json:"brand"`
	Code            *string `json:"code"`
	IsCheckNickname *string `json:"is_check_nickname"`
	Status          *string `json:"status"`
	Thumbnail       *string `json:"thumbnail"`
	Type            *string `json:"type"`
	Instruction     *string `json:"instruction"`
	Information     *string `json:"information"`
}
