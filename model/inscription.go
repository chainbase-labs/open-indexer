package model

type Inscription struct {
	Id          string `gorm:"column:id;primaryKey"`
	Number      uint64 `gorm:"column:number"`
	From        string `gorm:"column:from"`
	To          string `gorm:"column:to"`
	Block       uint64 `gorm:"column:block"`
	Idx         uint32 `gorm:"column:idx"`
	Timestamp   uint64 `gorm:"column:timestamp"`
	ContentType string `gorm:"column:content_type"`
	Content     string `gorm:"column:content"`
}
