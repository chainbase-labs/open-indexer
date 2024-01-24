package handlers

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"open-indexer/model"
	"open-indexer/utils"
	"sort"
	"strings"
	"time"
)

// The following data is only stored in memory, in practice it should be stored in a database such as mysql or mongodb
var inscriptions []*model.Inscription
var logEvents []*model.EvmLog
var asc20Records []*model.Asc20
var tokens = make(map[string]*model.Token)
var userBalances = make(map[string]map[string]*model.DDecimal)
var tokenHolders = make(map[string]map[string]*model.DDecimal)
var tokensByHash = make(map[string]*model.Token)
var lists = make(map[string]*model.List)
var tokenBalances = make(map[string]map[string]*model.TokenBalance)
var inscriptionNumber uint64 = 0

func GetInfo() (map[string]*model.Token, map[string]map[string]*model.DDecimal, map[string]map[string]*model.DDecimal) {
	return tokens, userBalances, tokenHolders
}

func GetTokenBalances() map[string]map[string]*model.TokenBalance {
	return tokenBalances
}

func GetTokenInfo() map[string]*model.Token {
	return tokens
}

func GetAsc20() []*model.Asc20 {
	return asc20Records
}

func GetInscriptions() []*model.Inscription {
	return inscriptions
}

func GetLogEvents() []*model.EvmLog {
	return logEvents
}

func SetTokens(tokeninfos []*model.Token) {
	for _, token := range tokeninfos {
		tokens[strings.ToLower(token.Tick)] = token
		tokenHolders[strings.ToLower(token.Tick)] = make(map[string]*model.DDecimal)
		tokensByHash[token.Hash] = token
	}
}

func SetLists(listList []*model.List) {
	for _, list := range listList {
		lists[list.InsId] = list
	}
}

func SetTokenBalances(tokenBalances []*model.TokenBalance) error {
	for _, balance := range tokenBalances {
		_, err := addBalance(balance.WalletAddress, balance.Tick, balance.Amount, balance.BlockNumber, uint64(balance.BlockTimestamp.Unix()))
		if err != nil {
			return err
		}
	}
	return nil
}

func MixRecords(trxs []*model.Transaction, logs []*model.EvmLog) []*model.Record {
	var records []*model.Record
	for _, trx := range trxs {
		var record model.Record
		record.IsLog = false
		record.Transaction = trx
		record.Block = trx.Block
		record.TransactionIndex = trx.Idx
		record.LogIndex = 0
		records = append(records, &record)
	}
	for _, log := range logs {
		var record model.Record
		record.IsLog = true
		record.EvmLog = log
		record.Block = log.Block
		record.TransactionIndex = log.TrxIndex
		record.LogIndex = log.LogIndex
		records = append(records, &record)
	}
	// resort
	sort.SliceStable(records, func(i, j int) bool {
		record0 := records[i]
		record1 := records[j]
		if record0.Block == record1.Block {
			if record0.TransactionIndex == record1.TransactionIndex {
				return record0.LogIndex+utils.BoolToUint32(record0.IsLog) < record1.LogIndex+utils.BoolToUint32(record1.IsLog)
			}
			return record0.TransactionIndex < record1.TransactionIndex
		}
		return record0.Block < record1.Block
	})
	return records
}

func ProcessRecords(records []*model.Record) error {
	logger.Println("records", len(records))
	var err error
	for _, record := range records {
		if record.IsLog {
			err = indexLog(record.EvmLog)
		} else {
			err = indexTransaction(record.Transaction)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func indexTransaction(trx *model.Transaction) error {
	if trx.ReceiptStatus != 1 {
		//logger.Warn("transaction invalid at block ", trx.Block, ":", trx.Idx)
		return nil
	}
	// data:,
	if !strings.HasPrefix(trx.Input, "0x646174613a") {
		return nil
	}
	bytes, err := hex.DecodeString(trx.Input[2:])
	if err != nil {
		logger.Warn("inscribe err", err, " at block ", trx.Block, ":", trx.Idx)
		return nil
	}
	input := string(bytes)

	sepIdx := strings.Index(input, ",")
	if sepIdx == -1 || sepIdx == len(input)-1 {
		return nil
	}
	contentType := "text/plain"
	if sepIdx > 5 {
		contentType = input[5:sepIdx]
	}
	content := input[sepIdx+1:]

	// save inscription
	inscriptionNumber++
	var inscription model.Inscription
	inscription.Number = inscriptionNumber
	inscription.Id = trx.Id
	inscription.From = trx.From
	inscription.To = trx.To
	inscription.Block = trx.Block
	inscription.Idx = trx.Idx
	inscription.Timestamp = trx.Timestamp
	inscription.ContentType = contentType
	inscription.Content = content

	if trx.To != "" {
		if err := handleProtocols(&inscription); err != nil {
			logger.Info("error at ", inscription.Number)
			return err
		}
	}

	inscriptions = append(inscriptions, &inscription)

	return nil
}

func indexLog(log *model.EvmLog) error {
	if log.Topic0 == "" {
		return nil
	}
	var topicType uint8
	if log.Topic0 == "0x8cdf9e10a7b20e7a9c4e778fc3eb28f2766e438a9856a62eac39fbd2be98cbc2" {
		// avascriptions_protocol_TransferASC20Token(address,address,string,uint256)
		topicType = 1
	} else if log.Topic0 == "0xe2750d6418e3719830794d3db788aa72febcd657bcd18ed8f1facdbf61a69a9a" {
		// avascriptions_protocol_TransferASC20TokenForListing(address,address,bytes32)
		topicType = 2
	} else {
		return nil
	}

	var asc20 model.Asc20
	asc20.Operation = "transfer"
	asc20.From = utils.TopicToAddress(log.Topic1)
	asc20.To = utils.TopicToAddress(log.Topic2)
	asc20.Block = log.Block
	asc20.Timestamp = log.Timestamp
	asc20.Hash = log.Hash
	asc20.LogIndex = log.LogIndex
	asc20.TrxIndex = log.TrxIndex
	if topicType == 1 {
		// transfer
		if asc20.From == log.Address {
			token, ok := tokensByHash[log.Topic3[2:]]
			if ok {
				asc20.Tick = token.Tick

				var err error
				asc20.Amount, asc20.Precision, err = model.NewDecimalFromString(utils.TopicToBigInt(log.Data).String())
				if err != nil {
					asc20.Valid = -56
				} else {
					// do transfer
					asc20.Valid, err = _transferToken(&asc20)
					if err != nil {
						return err
					}
				}
			} else {
				asc20.Valid = -51
			}
		} else {
			asc20.Valid = -52
			logger.Warningln("failed to validate transfer from:", asc20.From, "address:", log.Address)
		}
	} else {
		// exchange
		asc20.Operation = "exchange"

		list, ok := lists[log.Data]
		if ok {
			if list.Owner == asc20.From && list.Exchange == log.Address {
				asc20.Tick = list.Tick
				asc20.Amount = list.Amount
				asc20.Precision = list.Precision

				// update from to exchange
				asc20.From = log.Address

				// do transfer
				var err error
				asc20.Valid, err = exchangeToken(list, asc20.To, asc20.Block, asc20.Timestamp)
				if err != nil {
					return err
				}
			} else {
				if list.Owner != asc20.From {
					asc20.Valid = -54
					logger.Warningln("failed to validate transfer from:", asc20.From, list.Owner)
				} else {
					asc20.Valid = -55
					logger.Warningln("failed to validate exchange:", log.Address, list.Exchange)
				}
			}

		} else {
			asc20.Valid = -53
			logger.Warningln("failed to transfer, list not found, id:", log.Data)
		}
	}

	// save asc20 record
	asc20Records = append(asc20Records, &asc20)

	// save log
	logEvents = append(logEvents, log)
	return nil
}

func handleProtocols(inscription *model.Inscription) error {
	content := strings.TrimSpace(inscription.Content)
	if len(content) == 0 {
		return nil
	}
	if content[0] == '{' {
		var protoData map[string]string
		err := json.Unmarshal([]byte(content), &protoData)
		if err != nil {
			return nil
			//logger.Info("json parse error: ", err, ", at ", inscription.Number)
		} else {
			value, ok := protoData["p"]
			if ok && strings.TrimSpace(value) != "" {
				protocol := strings.ToLower(value)
				if protocol == "asc-20" {
					var asc20 model.Asc20
					asc20.Number = inscription.Number
					asc20.From = inscription.From
					asc20.To = inscription.To
					asc20.Block = inscription.Block
					asc20.Timestamp = inscription.Timestamp
					asc20.Hash = inscription.Id
					if value, ok = protoData["tick"]; ok {
						asc20.Tick = strings.ToLower(value)
					}
					if value, ok = protoData["op"]; ok {
						asc20.Operation = value
					}

					var err error
					if strings.TrimSpace(asc20.Tick) == "" {
						asc20.Valid = -1 // empty tick
					} else if len(asc20.Tick) > 18 {
						asc20.Valid = -2 // too long tick
					} else if asc20.Operation == "deploy" {
						asc20.Valid, err = deployToken(&asc20, protoData)
					} else if asc20.Operation == "mint" {
						asc20.Valid, err = mintToken(&asc20, protoData)
					} else if asc20.Operation == "transfer" {
						asc20.Valid, err = transferToken(&asc20, protoData)
					} else if asc20.Operation == "list" {
						asc20.Valid, err = listToken(&asc20, protoData)
					} else {
						asc20.Valid = -3 // wrong operation
					}
					if err != nil {
						return err
					}

					// save asc20 records
					asc20Records = append(asc20Records, &asc20)
					return nil
				}
			}
		}
	}
	return nil
}

func deployToken(asc20 *model.Asc20, params map[string]string) (int8, error) {
	value, ok := params["max"]
	if !ok {
		return -11, nil
	}
	max, precision, err1 := model.NewDecimalFromString(value)
	if err1 != nil {
		return -12, nil
	}
	if precision != 0 {
		// Currently only 0 precision is supported
		return -12, nil
	}
	value, ok = params["lim"]
	if !ok {
		return -13, nil
	}
	limit, _, err2 := model.NewDecimalFromString(value)
	if err2 != nil {
		return -14, nil
	}
	if max.Sign() <= 0 || limit.Sign() <= 0 {
		return -15, nil
	}
	if utils.ParseInt64(max.String()) == 0 || utils.ParseInt64(limit.String()) == 0 {
		return -15, nil
	}
	if max.Cmp(limit) < 0 {
		return -16, nil
	}

	asc20.Max = max
	asc20.Precision = precision
	asc20.Limit = limit

	// 已经 deploy
	asc20.Tick = strings.TrimSpace(asc20.Tick) // trim tick
	_, exists := tokens[strings.ToLower(asc20.Tick)]
	if exists {
		logger.Info("token ", asc20.Tick, " has deployed at ", asc20.Number)
		return -17, nil
	}

	token := &model.Token{
		Tick:               asc20.Tick,
		Number:             asc20.Number,
		Precision:          precision,
		Max:                max,
		Limit:              limit,
		Minted:             model.NewDecimal(),
		Progress:           0,
		CreatedAt:          asc20.Timestamp,
		CompletedAt:        0,
		Hash:               utils.Keccak256(strings.ToLower(asc20.Tick)),
		TxHash:             asc20.Hash,
		Creator:            asc20.From,
		TxIndex:            asc20.TrxIndex,
		CreatedBlockNumber: asc20.Block,
		UpdatedAt:          asc20.Timestamp,
	}

	// save
	tokens[strings.ToLower(token.Tick)] = token
	tokenHolders[strings.ToLower(token.Tick)] = make(map[string]*model.DDecimal)
	tokensByHash[token.Hash] = token

	return 1, nil
}

func mintToken(asc20 *model.Asc20, params map[string]string) (int8, error) {
	value, ok := params["amt"]
	if !ok {
		return -21, nil
	}
	amt, precision, err := model.NewDecimalFromString(value)
	if err != nil {
		return -22, nil
	}

	asc20.Amount = amt

	// check token
	tick := strings.ToLower(asc20.Tick)
	token, exists := tokens[tick]
	if !exists {
		return -23, nil
	}
	asc20.Tick = token.Tick

	// check precision
	if precision > token.Precision {
		return -24, nil
	}

	if amt.Sign() <= 0 {
		return -25, nil
	}

	if amt.Cmp(token.Limit) == 1 {
		return -26, nil
	}

	var left = token.Max.Sub(token.Minted)

	if left.Cmp(amt) == -1 {
		if left.Sign() > 0 {
			amt = left
		} else {
			// exceed max
			return -27, nil
		}
	}
	// update amount
	asc20.Amount = amt
	asc20.Precision = precision

	newHolder, err := addBalance(asc20.To, tick, amt, asc20.Block, asc20.Timestamp)
	if err != nil {
		return 0, err
	}

	// update token
	token.Minted = token.Minted.Add(amt)
	token.Trxs++

	if token.Minted.Cmp(token.Max) >= 0 {
		token.Progress = 1000000
	} else {
		token.Progress = int32(utils.ParseInt64(token.Minted.String()) * 1000000 / utils.ParseInt64(token.Max.String()))
	}

	if token.Minted.Cmp(token.Max) == 0 {
		token.CompletedAt = asc20.Timestamp
	}
	if newHolder {
		token.Holders++
	}
	token.UpdatedAt = asc20.Timestamp

	return 1, err
}

func transferToken(asc20 *model.Asc20, params map[string]string) (int8, error) {
	value, ok := params["amt"]
	if !ok {
		return -31, nil
	}
	amt, precision, err := model.NewDecimalFromString(value)
	if err != nil {
		return -32, nil
	}

	asc20.Amount = amt
	asc20.Precision = precision

	return _transferToken(asc20)
}

func listToken(asc20 *model.Asc20, params map[string]string) (int8, error) {
	value, ok := params["amt"]
	if !ok {
		return -31, nil
	}
	amt, precision, err := model.NewDecimalFromString(value)
	if err != nil {
		return -32, nil
	}

	// check token
	tick := strings.ToLower(asc20.Tick)
	token, exists := tokens[tick]
	if !exists {
		return -33, nil
	}
	asc20.Tick = token.Tick

	// check precision
	if precision > token.Precision {
		return -34, nil
	}

	if amt.Sign() <= 0 {
		return -35, nil
	}

	if asc20.From == asc20.To {
		// list to self
		return -36, nil
	}

	asc20.Amount = amt

	// sub balance
	reduceHolder, err := subBalance(asc20.From, tick, amt, asc20.Block, asc20.Timestamp)
	if err != nil {
		if err.Error() == "insufficient balance" {
			return -37, nil
		}
		return 0, err
	}

	// add list
	var list model.List
	list.InsId = asc20.Hash
	list.Owner = asc20.From
	list.Exchange = asc20.To
	list.Tick = token.Tick
	list.Amount = amt
	list.Precision = precision

	lists[list.InsId] = &list

	token.Trxs++

	if reduceHolder {
		token.Holders--
	}

	return 1, err
}

func exchangeToken(list *model.List, sendTo string, number uint64, timestamp uint64) (int8, error) {

	// add balance
	newHolder, err := addBalance(sendTo, list.Tick, list.Amount, number, timestamp)
	if err != nil {
		return 0, err
	}

	// update token
	tick := strings.ToLower(list.Tick)
	token, exists := tokens[tick]
	if !exists {
		return -33, nil
	}

	token.Trxs++

	if newHolder {
		token.Holders++
	}
	token.UpdatedAt = timestamp

	// delete list from lists
	delete(lists, list.InsId)
	logger.Println("exchange", list.Amount)
	return 1, err
}

func _transferToken(asc20 *model.Asc20) (int8, error) {

	// check token
	tick := strings.ToLower(asc20.Tick)
	token, exists := tokens[tick]
	if !exists {
		return -33, nil
	}
	asc20.Tick = token.Tick

	if asc20.Precision > token.Precision {
		return -34, nil
	}

	if asc20.Amount.Sign() <= 0 {
		return -35, nil
	}

	if asc20.From == "" || asc20.To == "" {
		// send to self
		return -9, nil
	}
	if asc20.From == asc20.To {
		// send to self
		return -36, nil
	}

	logger.Infof("tx hash %s", asc20.Hash)
	// From
	reduceHolder, err := subBalance(asc20.From, tick, asc20.Amount, asc20.Block, asc20.Timestamp)
	if err != nil {
		if err.Error() == "insufficient balance" {
			return -37, nil
		}
		return 0, err
	}

	// To
	newHolder, err := addBalance(asc20.To, tick, asc20.Amount, asc20.Block, asc20.Timestamp)
	if err != nil {
		return 0, err
	}

	// update token
	if reduceHolder {
		token.Holders--
	}
	if newHolder {
		token.Holders++
	}
	token.Trxs++
	token.UpdatedAt = asc20.Timestamp

	return 1, err
}

func subBalance(owner string, tick string, amount *model.DDecimal, number uint64, timestamp uint64) (bool, error) {
	token, exists := tokens[strings.ToLower(tick)]
	if !exists {
		return false, errors.New("token not found")
	}
	fromBalances, ok := userBalances[owner]
	if !ok {
		return false, errors.New("insufficient balance")
	}
	fromBalance, ok := fromBalances[token.Tick]
	logger.Infof("sub balance from %s ,tick %s,frombalance %s, amount %s, blockNumber %d", owner, tick, fromBalance.String(), amount.String(), number)
	if amount.Cmp(fromBalance) == 1 {
		return false, errors.New("insufficient balance")
	}

	fromBalance = fromBalance.Sub(amount)

	var reduceHolder = false
	if fromBalance.Sign() == 0 {
		reduceHolder = true
	}

	// save
	fromBalances[token.Tick] = fromBalance
	tokenHolders[token.Tick][owner] = fromBalance

	tickHolders, ok := tokenBalances[token.Tick]
	if !ok {
		tickHolders = make(map[string]*model.TokenBalance)
		tokenBalances[tick] = tickHolders
	}

	ownerBalance, ok := tickHolders[owner]
	if !ok {
		tickHolders[owner] = &model.TokenBalance{
			BlockNumber:    number,
			BlockTimestamp: time.Unix(int64(timestamp), 0),
			Tick:           tick,
			WalletAddress:  owner,
			TotalSupply:    token.Max,
			Amount:         fromBalance,
		}
	} else {
		ownerBalance.Amount = fromBalance
		ownerBalance.Amount = fromBalance
		ownerBalance.BlockNumber = number
		ownerBalance.BlockTimestamp = time.Unix(int64(timestamp), 0)
	}

	return reduceHolder, nil
}

func addBalance(owner string, tick string, amount *model.DDecimal, number uint64, timestamp uint64) (bool, error) {
	token, exists := tokens[strings.ToLower(tick)]
	if !exists {
		return false, errors.New("token not found")
	}
	toBalances, ok := userBalances[owner]
	if !ok {
		toBalances = make(map[string]*model.DDecimal)
		userBalances[owner] = toBalances
	}
	var newHolder = false
	toBalance, ok := toBalances[token.Tick]
	if !ok {
		toBalance = model.NewDecimal()
		newHolder = true
	}

	toBalance = toBalance.Add(amount)

	if toBalance.Sign() == 0 {
		newHolder = true
	}

	// save
	toBalances[token.Tick] = toBalance
	tokenHolders[token.Tick][owner] = toBalance

	tickHolders, ok := tokenBalances[token.Tick]
	if !ok {
		tickHolders = make(map[string]*model.TokenBalance)
		tokenBalances[tick] = tickHolders
	}

	ownerBalance, ok := tickHolders[owner]
	if !ok {
		tickHolders[owner] = &model.TokenBalance{
			BlockNumber:    number,
			BlockTimestamp: time.Unix(int64(timestamp), 0),
			Tick:           tick,
			WalletAddress:  owner,
			TotalSupply:    token.Max,
			Amount:         toBalance,
		}
	} else {
		ownerBalance.Amount = toBalance
		ownerBalance.BlockNumber = number
		ownerBalance.BlockTimestamp = time.Unix(int64(timestamp), 0)
	}

	return newHolder, nil
}
