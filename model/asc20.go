package model

type Asc20 struct {
	Number    uint64
	Tick      string
	From      string
	To        string
	Operation string
	Precision int
	Max       *DDecimal
	Limit     *DDecimal
	Amount    *DDecimal
	Hash      string
	Block     uint64
	Timestamp uint64
	TrxIndex  uint32
	LogIndex  uint32
	Valid     int8
}
