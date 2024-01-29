package model

import (
	"time"
)

type TokenInfoHis struct {
	BlockTimestamp   time.Time  `gorm:"column:block_timestamp"`
	BlockNumber      uint64     `gorm:"column:block_number;primaryKey"`
	TxIndex          uint32     `gorm:"column:tx_index"`
	TxHash           string     `gorm:"column:tx_hash"`
	Tick             string     `gorm:"column:tick;primaryKey"`
	MaxSupply        *DDecimal  `gorm:"column:max_supply"`
	Lim              *DDecimal  `gorm:"column:lim"`
	Wlim             *DDecimal  `gorm:"column:wlim"`
	Dec              int        `gorm:"column:dec"`
	Creator          string     `gorm:"column:creator"`
	Minted           *DDecimal  `gorm:"column:minted"`
	Holders          int32      `gorm:"column:holders"`
	Txs              int32      `gorm:"column:txs"`
	UpdatedTimeStamp time.Time  `gorm:"column:updated_timestamp"`
	CompletedAt      *time.Time `gorm:"column:completed_timestamp"`
	ID               string     `gorm:"column:id"`
}

func (TokenInfoHis) TableName() string {
	return "token_info_his"
}
