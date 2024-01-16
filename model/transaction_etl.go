package model

type TransactionETL struct {
	Hash                     string `gorm:"primaryKey"`
	Nonce                    uint64
	TransactionIndex         uint32
	FromAddress              string
	ToAddress                string
	Value                    string
	Gas                      string
	GasPrice                 string
	Input                    string
	ReceiptCumulativeGasUsed string
	ReceiptGasUsed           string
	ReceiptContractAddress   string
	ReceiptRoot              string
	ReceiptStatus            uint8
	BlockTimestamp           uint64
	BlockNumber              uint64
	BlockHash                string
	MaxFeePerGas             string
	MaxPriorityFeePerGas     string
	TransactionType          uint64
	ReceiptEffectiveGasPrice string
}
