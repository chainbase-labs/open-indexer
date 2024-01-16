package model

import (
	"time"
)

type Raw struct {
	BlockTimestamp time.Time `gorm:"column:block_timestamp"`
	BlockNumber    int64     `gorm:"column:block_number"`
	TxIndex        int       `gorm:"column:tx_index"`
	TxHash         string    `gorm:"column:tx_hash"`
	Nonce          string    `gorm:"column:nonce"`
	LogIndex       int       `gorm:"column:log_index"`
	P              string    `gorm:"column:p"`
	Op             string    `gorm:"column:op"`
	Tick           string    `gorm:"column:tick"`
	MaxSupply      string    `gorm:"column:max_supply"`
	Lim            string    `gorm:"column:lim"`
	Wlim           string    `gorm:"column:wlim"`
	Dec            int       `gorm:"column:dec"`
	ID             string    `gorm:"column:id"`
	Amt            string    `gorm:"column:amt"`
	EthsFrom       string    `gorm:"column:eths_from"`
	EthsNonce      string    `gorm:"column:eths_nonce"`
	Creator        string    `gorm:"column:creator"`
	Owner          string    `gorm:"column:owner"`
	DataContent    string    `gorm:"column:data_content"`
	Data           string    `gorm:"column:data"`
	PK             int       `gorm:"column:__pk"`
}
