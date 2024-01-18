package model

type EvmLog struct {
	Hash      string `gorm:"column:hash;primaryKey"`
	Address   string `gorm:"column:address"`
	Topic0    string `gorm:"column:topic0"`
	Topic1    string `gorm:"column:topic1"`
	Topic2    string `gorm:"column:topic2"`
	Topic3    string `gorm:"column:topic3"`
	Data      string `gorm:"column:data"`
	Block     uint64 `gorm:"column:block_number"`
	TrxIndex  uint32 `gorm:"column:tx_index"`
	LogIndex  uint32 `gorm:"column:log_index;primaryKey"`
	Timestamp uint64 `gorm:"column:number"`
}

func (EvmLog) TableName() string {
	return "evm_logs"
}

//func NewEvmLogFromMixData(mixData *MixData) *EvmLog {
//	logEvent := mixData.LogEvent
//	topic0 := ""
//	topic1 := ""
//	topic2 := ""
//	topic3 := ""
//	if len(logEvent.Topics) > 0 {
//		topic0 = logEvent.Topics[0]
//		if len(logEvent.Topics) > 1 {
//			topic1 = logEvent.Topics[1]
//			if len(logEvent.Topics) > 2 {
//				topic2 = logEvent.Topics[2]
//				if len(logEvent.Topics) > 3 {
//					topic3 = logEvent.Topics[3]
//				}
//			}
//		}
//	}
//	var log = EvmLog{
//		Hash:      mixData.Transaction.Hash,
//		Address:   logEvent.Address,
//		Topic0:    topic0,
//		Topic1:    topic1,
//		Topic2:    topic2,
//		Topic3:    topic3,
//		Data:      logEvent.Data,
//		Block:     mixData.BlockNumber,
//		TrxIndex:  mixData.TransactionIndex,
//		LogIndex:  mixData.LogIndex,
//		Timestamp: mixData.TimeStamp,
//		Type:      0,
//		Status:    0,
//	}
//
//	return &log
//}
