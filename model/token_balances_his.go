package model

import "time"

type TokenBalanceHis struct {
	BlockNumber    uint64    `gorm:"column:block_number;primaryKey"`
	BlockTimestamp time.Time `gorm:"column:block_timestamp;type:datetime(3)"`
	Tick           string    `gorm:"column:tick;primaryKey"`
	WalletAddress  string    `gorm:"column:wallet_address;primaryKey"`
	TotalSupply    *DDecimal `gorm:"column:total_supply;type:decimal(38,0)"`
	Amount         *DDecimal `gorm:"column:amount;type:decimal(38,0)"`
}

func (TokenBalanceHis) TableName() string {
	return "token_balances_his"
}
