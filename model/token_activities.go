package model

import "time"

type TokenActivity struct {
	BlockTimestamp time.Time `gorm:"column:block_timestamp;primaryKey"`
	BlockNumber    uint64    `gorm:"column:block_number;primaryKey"`
	TxIndex        uint32    `gorm:"column:tx_index;primaryKey"`
	TxHash         string    `gorm:"column:tx_hash"`
	LogIndex       uint32    `gorm:"column:log_index;primaryKey"`
	Type           string    `gorm:"column:type;primaryKey"`
	Tick           string    `gorm:"column:tick;primaryKey"`
	ID             string    `gorm:"column:id;primaryKey"`
	Amt            *DDecimal `gorm:"column:amt"`
	FromAddress    string    `gorm:"column:from_address"`
	ToAddress      string    `gorm:"column:to_address"`
}
