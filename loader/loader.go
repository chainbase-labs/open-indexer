package loader

import (
	"bufio"
	"fmt"
	"gorm.io/gorm"
	"log"
	"open-indexer/connector/tidb"
	"open-indexer/model"
	"open-indexer/utils"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

var maxBlockNumber uint64 = 0

type Holder struct {
	Address string
	Amount  *model.DDecimal
}

func LoadTransactionData(fname string) ([]*model.Transaction, error) {
	file, err := os.Open(fname)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var trxs []*model.Transaction
	scanner := bufio.NewScanner(file)
	max := 4 * 1024 * 1024
	buf := make([]byte, max)
	scanner.Buffer(buf, max)

	for scanner.Scan() {
		line := scanner.Text()
		//log.Printf(line)
		fields := strings.Split(line, ",")

		if len(fields) != 21 {
			return nil, fmt.Errorf("invalid data format", len(fields), ":", fields)
		}

		block, err := strconv.ParseUint(fields[15], 10, 32)
		if err != nil {
			return nil, err
		}

		if block <= maxBlockNumber {
			continue
		}

		var data model.TransactionETL

		data.Hash = fields[0]
		nonce, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			return nil, err
		}
		data.Nonce = nonce

		txridx, err := strconv.ParseUint(fields[2], 10, 64)
		if err != nil {
			return nil, err
		}
		data.TransactionIndex = uint32(txridx)

		data.FromAddress = fields[3]
		data.ToAddress = fields[4]
		data.Value = fields[5]
		data.Gas = fields[6]
		data.GasPrice = fields[7]
		data.Input = fields[8]
		data.ReceiptCumulativeGasUsed = fields[9]
		data.ReceiptGasUsed = fields[10]

		data.ReceiptContractAddress = fields[11]
		data.ReceiptRoot = fields[12]

		recStatus, err := strconv.ParseUint(fields[13], 10, 64)
		if err != nil {
			return nil, err
		}
		data.ReceiptStatus = uint8(recStatus)

		blockTime, err := strconv.ParseUint(fields[14], 10, 32)
		if err != nil {
			return nil, err
		}
		data.BlockTimestamp = blockTime

		data.BlockNumber = block

		data.BlockHash = fields[16]
		data.MaxFeePerGas = fields[17]
		data.MaxPriorityFeePerGas = fields[18]

		txrType, err := strconv.ParseUint(fields[19], 10, 32)
		if err != nil {
			return nil, err
		}
		data.TransactionType = txrType

		data.ReceiptEffectiveGasPrice = fields[20]

		dataEtl := ConvertTransactionETLToTransaction(data)

		trxs = append(trxs, &dataEtl)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return trxs, nil
}

func LoadLogData(fname string) ([]*model.EvmLog, error) {

	file, err := os.Open(fname)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var logs []*model.EvmLog
	scanner := bufio.NewScanner(file)
	max := 4 * 1024 * 1024
	buf := make([]byte, max)
	scanner.Buffer(buf, max)

	for scanner.Scan() {
		line := scanner.Text()
		//log.Printf(line)
		fields := strings.Split(line, ",")

		if len(fields) != 12 {
			return nil, fmt.Errorf("invalid data format", len(fields))
		}

		blockNumber, err := strconv.ParseUint(fields[10], 10, 64)
		if err != nil {
			return nil, err
		}
		if blockNumber <= maxBlockNumber {
			continue
		}
		var log model.EvmLogETL

		logIdx, err := strconv.ParseUint(fields[0], 10, 32)
		if err != nil {
			return nil, err
		}
		log.LogIndex = uint32(logIdx)

		log.TransactionHash = fields[1]

		trxIdx, err := strconv.ParseUint(fields[2], 10, 32)
		if err != nil {
			return nil, err
		}
		log.TransactionIndex = uint32(trxIdx)

		log.Address = fields[3]
		log.Data = fields[4]
		log.Topic0 = fields[5]
		log.Topic1 = fields[6]
		log.Topic2 = fields[7]
		log.Topic3 = fields[8]

		blockTimestamp, err := strconv.ParseUint(fields[9], 10, 64)
		if err != nil {
			return nil, err
		}
		log.BlockTimestamp = blockTimestamp

		log.BlockNumber = blockNumber
		log.BlockHash = fields[11]

		var evmLog = ConvertEvmLogETLToEvmLog(log)

		logs = append(logs, &evmLog)
	}

	return logs, nil
}

func DumpTickerInfoMap(fname string,
	tokens map[string]*model.Token,
	userBalances map[string]map[string]*model.DDecimal,
	tokenHolders map[string]map[string]*model.DDecimal,
) {

	file, err := os.OpenFile(fname, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0777)
	if err != nil {
		log.Fatalf("open block index file failed, %s", err)
		return
	}
	defer file.Close()

	var allTickers []string
	for ticker := range tokens {
		allTickers = append(allTickers, ticker)
	}
	sort.SliceStable(allTickers, func(i, j int) bool {
		return allTickers[i] < allTickers[j]
	})

	for _, ticker := range allTickers {
		info := tokens[ticker]

		fmt.Fprintf(file, "%s trxs: %d, total: %s, minted: %s, holders: %d\n",
			info.Tick,
			info.Trxs,
			info.Max.String(),
			info.Minted,
			len(tokenHolders[ticker]),
		)

		// holders
		var allHolders []Holder
		for address := range tokenHolders[ticker] {
			holder := Holder{
				address,
				tokenHolders[ticker][address],
			}
			allHolders = append(allHolders, holder)
		}

		sort.SliceStable(allHolders, func(i, j int) bool {
			return allHolders[i].Amount.Cmp(allHolders[j].Amount) > 0
		})

		// holders
		for _, holder := range allHolders {

			fmt.Fprintf(file, "%s %s  balance: %s\n",
				info.Tick,
				holder.Address,
				holder.Amount.String(),
			)
		}
	}
}

func ConvertEvmLogETLToEvmLog(etlLog model.EvmLogETL) model.EvmLog {
	return model.EvmLog{
		Hash:      etlLog.TransactionHash,
		Address:   etlLog.Address,
		Topic0:    etlLog.Topic0,
		Topic1:    etlLog.Topic1,
		Topic2:    etlLog.Topic2,
		Topic3:    etlLog.Topic3,
		Data:      etlLog.Data,
		Block:     etlLog.BlockNumber,
		TrxIndex:  etlLog.TransactionIndex,
		LogIndex:  etlLog.LogIndex,
		Timestamp: etlLog.BlockTimestamp,
	}
}

func ConvertTransactionETLToTransaction(etl model.TransactionETL) model.Transaction {
	return model.Transaction{
		Id:            etl.Hash,
		From:          etl.FromAddress,
		To:            etl.ToAddress,
		Block:         etl.BlockNumber,
		Idx:           etl.TransactionIndex,
		Timestamp:     etl.BlockTimestamp,
		Input:         etl.Input,
		ReceiptStatus: int8(etl.ReceiptStatus),
	}
}

func ConvertTokensToTokenInfos(tokens map[string]*model.Token) []*model.TokenInfo {
	var tokenInfos []*model.TokenInfo
	for _, token := range tokens {
		tokenInfo := &model.TokenInfo{
			BlockTimestamp:   time.Unix(int64(token.CreatedAt), 0),
			BlockNumber:      token.CreatedBlockNumber,
			ID:               strconv.FormatUint(token.Number, 10),
			TxIndex:          token.TxIndex,
			TxHash:           token.TxHash,
			Tick:             token.Tick,
			MaxSupply:        token.Max,
			Lim:              token.Limit,
			Wlim:             nil,
			Dec:              token.Precision,
			Creator:          token.Creator,
			Minted:           token.Minted,
			Holders:          token.Holders,
			Txs:              token.Trxs,
			UpdatedTimeStamp: time.Unix(int64(token.UpdatedAt), 0),
		}
		if token.CompletedAt != 0 {
			t := time.Unix(int64(token.CompletedAt), 0)
			tokenInfo.CompletedAt = &t
		}
		tokenInfos = append(tokenInfos, tokenInfo)
	}

	return tokenInfos
}

func ConvertAsc20sToTokenActivities(asc20s []*model.Asc20) []*model.TokenActivity {
	var tokenActivities []*model.TokenActivity
	for _, asc20 := range asc20s {
		if (asc20.Operation == "list" || asc20.Operation == "mint" || asc20.Operation == "transfer" || asc20.Operation == "exchange") && asc20.Valid == 1 {
			tokenActivity := &model.TokenActivity{
				BlockTimestamp: time.Unix(int64(asc20.Timestamp), 0),
				BlockNumber:    asc20.Block,
				TxIndex:        asc20.TrxIndex,
				TxHash:         asc20.Hash,
				LogIndex:       asc20.LogIndex,
				Type:           asc20.Operation,
				Tick:           asc20.Tick,
				ID:             strconv.FormatUint(asc20.Number, 10),
				Amt:            asc20.Amount,
				FromAddress:    asc20.From,
				ToAddress:      asc20.To,
			}
			tokenActivities = append(tokenActivities, tokenActivity)
		}
	}
	return tokenActivities
}

func LoadTokenInfo(db *gorm.DB) ([]*model.Token, error) {
	var tokenInfos []*model.TokenInfo
	var tokens []*model.Token
	tableName := model.TokenInfo{}.TableName()
	exist, err := tidb.JudgeTableExistOrNot(db, tableName)
	if !exist {
		return tokens, nil
	}

	if err != nil {
		return tokens, err
	}

	err = db.Find(&tokenInfos).Error
	if err != nil {
		return tokens, err
	}

	for _, tokenInfo := range tokenInfos {
		token, _ := ConvertTokenInfoToToken(tokenInfo)
		tokens = append(tokens, token)
	}
	return tokens, nil
}

func ConvertTokenInfoToToken(tokenInfo *model.TokenInfo) (*model.Token, error) {
	var completedAt uint64
	if tokenInfo.CompletedAt == nil {
		completedAt = 0
	} else {
		completedAt = uint64(tokenInfo.CompletedAt.Unix())
	}

	var num uint64

	if tokenInfo.ID != "" {
		var err error
		num, err = strconv.ParseUint(tokenInfo.ID, 10, 64)
		if err != nil {
			return nil, err
		}
	}

	token := &model.Token{
		Tick:               tokenInfo.Tick,
		Number:             num,
		CreatedBlockNumber: tokenInfo.BlockNumber,
		Precision:          tokenInfo.Dec,
		Max:                tokenInfo.MaxSupply,
		Limit:              tokenInfo.Lim,
		Minted:             tokenInfo.Minted,
		Progress:           0,
		Holders:            tokenInfo.Holders,
		Trxs:               tokenInfo.Txs,
		CreatedAt:          uint64(tokenInfo.BlockTimestamp.Unix()),
		UpdatedAt:          uint64(tokenInfo.UpdatedTimeStamp.Unix()),
		CompletedAt:        completedAt,
		Hash:               utils.Keccak256(strings.ToLower(tokenInfo.Tick)),
		TxHash:             tokenInfo.TxHash,
		TxIndex:            tokenInfo.TxIndex,
		Creator:            tokenInfo.Creator,
	}
	return token, nil
}

func LoadTokenBalances(db *gorm.DB, rerun bool, rerun_start uint64) ([]*model.TokenBalance, error) {
	var tokenBalances []*model.TokenBalance

	if rerun {
		exist, err := tidb.JudgeTableExistOrNot(db, model.TokenBalanceHis{}.TableName())
		if !exist {
			return tokenBalances, nil
		}
		if err != nil {
			return tokenBalances, err
		}

		db.Where("block_number > ?", rerun_start).Delete(&model.TokenBalanceHis{})

		err = db.Raw(`
WITH RankedTokenBalances AS (
    SELECT *,
           ROW_NUMBER() OVER (PARTITION BY wallet_address, tick ORDER BY block_number DESC) as rn
    FROM token_balances_his
)
SELECT * 
FROM RankedTokenBalances
WHERE rn = 1
`).Scan(&tokenBalances).Error
		if err != nil {
			return tokenBalances, err
		}
	} else {
		exist, err := tidb.JudgeTableExistOrNot(db, model.TokenBalance{}.TableName())
		if !exist {
			return tokenBalances, nil
		}
		if err != nil {
			return tokenBalances, err
		}
		err = db.Find(&tokenBalances).Error
		if err != nil {
			return tokenBalances, err
		}
	}

	for _, tokenBalance := range tokenBalances {
		if tokenBalance.BlockNumber > maxBlockNumber {
			maxBlockNumber = tokenBalance.BlockNumber
		}
	}
	return tokenBalances, nil
}

func SetMaxBlockNumber(max uint64) {
	maxBlockNumber = max
}

func GetMaxBlockNumberFromDB(db *gorm.DB) error {
	exist, err := tidb.JudgeTableExistOrNot(db, model.Inscription{}.TableName())
	if exist {
		db.Raw("SELECT MAX(block) FROM inscriptions").Scan(&maxBlockNumber)
	}
	exist, err = tidb.JudgeTableExistOrNot(db, model.EvmLog{}.TableName())
	if exist {
		var maxBlockLogs *uint64
		db.Raw("SELECT MAX(block_number) FROM evm_logs").Scan(&maxBlockLogs)
		if maxBlockLogs != nil && maxBlockNumber < *maxBlockLogs {
			maxBlockNumber = *maxBlockLogs
		}
	}
	return err
}

func LoadList(db *gorm.DB) ([]*model.List, error) {
	var tokenActivities []*model.TokenActivity
	var lists []*model.List

	exist, err := tidb.JudgeTableExistOrNot(db, model.TokenActivity{}.TableName())
	if !exist {
		return lists, nil
	}
	if err != nil {
		return lists, err
	}

	err = db.Where("type = ?", "list").Find(&tokenActivities).Error
	if err != nil {
		return lists, err
	}

	for _, activity := range tokenActivities {
		if activity.Type == "list" {
			list, _ := ConvertTokenActivityToList(activity)
			lists = append(lists, list)
		}
	}
	return lists, nil
}

func ConvertTokenActivityToList(activity *model.TokenActivity) (*model.List, error) {
	amount, precision, err := model.NewDecimalFromString(activity.Amt.String())
	if err != nil {
		return nil, fmt.Errorf("error converting Amt to *DDecimal: %v", err)
	}

	list := &model.List{
		InsId:     activity.TxHash,
		Owner:     activity.FromAddress,
		Exchange:  activity.ToAddress,
		Tick:      activity.Tick,
		Amount:    amount,
		Precision: precision,
	}
	return list, nil
}
