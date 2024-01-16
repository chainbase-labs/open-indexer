package model

type EvmLogETL struct {
	LogIndex         uint32 `gorm:"primaryKey"`
	TransactionHash  string `gorm:"primaryKey"`
	TransactionIndex uint32
	Address          string
	Data             string
	Topic0           string
	Topic1           string
	Topic2           string
	Topic3           string
	BlockTimestamp   uint64
	BlockNumber      uint64
	BlockHash        string
}
