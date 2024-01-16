package model

type List struct {
	InsId     string `gorm:"primaryKey"`
	Owner     string
	Exchange  string
	Tick      string
	Amount    *DDecimal
	Precision int
}
