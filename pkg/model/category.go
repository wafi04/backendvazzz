package model

type Category struct {
	ID              int     `json:"id"`
	Name            string  `json:"name"`
	SubName         string  `json:"subName"`
	Brand           string  `json:"brand"`
	Code            string  `json:"code"`
	IsCheckNickname string  `json:"isCheckNickname"`
	Status          string  `json:"status"`
	Thumbnail       string  `json:"thumbnail"`
	Type            string  `json:"type"`
	Banner          string  `json:"banner"`
	Instruction     *string `json:"instruction,omitempty"`
	Information     *string `json:"information,omitempty"`
	Placeholder1    string  `json:"placeholder1"`
	Placeholder2    string  `json:"placeholder2"`
	CreatedAt       string  `json:"createdAt"`
	UpdatedAt       string  `json:"updatedAt"`
}

type CreateCategory struct {
	Name            string  `json:"name"`
	SubName         string  `json:"subName"`
	Brand           string  `json:"brand"`
	Code            string  `json:"code"`
	IsCheckNickname string  `json:"isCheckNickname"`
	Status          string  `json:"status"`
	Thumbnail       string  `json:"thumbnail"`
	Type            string  `json:"type"`
	Banner          string  `json:"banner"`
	Placeholder1    string  `json:"placeholder1"`
	Placeholder2    string  `json:"placeholder2"`
	Instruction     *string `json:"instruction,omitempty"`
	Information     *string `json:"information,omitempty"`
}

type UpdateCategory struct {
	Name            *string `json:"name"`
	SubName         *string `json:"subName"`
	Brand           *string `json:"brand"`
	Code            *string `json:"code"`
	IsCheckNickname *string `json:"isCheckNickname"`
	Status          *string `json:"status"`
	Thumbnail       *string `json:"thumbnail"`
	Type            *string `json:"type"`
	Instruction     *string `json:"instruction,omitempty"`
	Information     *string `json:"information,omitempty"`
}
