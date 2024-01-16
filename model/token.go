package model

type Token struct {
	Tick               string
	Number             uint64
	Precision          int
	Max                *DDecimal
	Limit              *DDecimal
	Minted             *DDecimal
	Progress           int32
	Holders            int32
	Trxs               int32
	CreatedAt          uint64
	CompletedAt        uint64
	UpdatedAt          uint64
	Hash               string
	TxHash             string
	TxIndex            uint32 `gorm:"column:tx_index"`
	Creator            string `gorm:"column:creator"`
	CreatedBlockNumber uint64
}
