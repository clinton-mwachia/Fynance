package models

type Report struct {
	Period       string  `gorm:"type:bigint;not null" json:"period"`
	TotalIncome  float64 `gorm:"type:bigint;not null" json:"total_income"`
	TotalExpense float64 `gorm:"type:bigint;not null" json:"total_expense"`
	Balance      float64 `gorm:"type:bigint;not null" json:"balance"`
}
