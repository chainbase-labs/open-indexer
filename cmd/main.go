package main

import (
	"flag"
	"log"
	"open-indexer/connector/tidb"
	"open-indexer/handlers"
	"open-indexer/loader"
)

var (
	inputfile1  string
	inputfile2  string
	rerun       bool
	rerun_start uint64
)

func init() {
	flag.StringVar(&inputfile1, "transactions", "", "the filename of input data, default(./data/transactions.input.txt)")
	flag.StringVar(&inputfile2, "logs", "", "the filename of input data, default(./data/logs.input.txt)")
	flag.BoolVar(&rerun, "rerun", false, "when rerun load token balance from token balance history, default(false)")
	flag.Uint64Var(&rerun_start, "rerun_start", 0, "when rerun delete token balance history that data > rerun_start")

	flag.Parse()

	if inputfile1 == "" || inputfile2 == "" {
		log.Fatal("Please specify both transaction and logs file paths.")
	}
}

func main() {

	var logger = handlers.GetLogger()

	logger.Info("start index")

	db, err := tidb.GetDBInstanceByEnv()

	tokenList, err := loader.LoadTokenInfo(db)
	if err != nil {
		logger.Fatalf("load token info failed, %s", err)
	}
	handlers.SetTokens(tokenList)

	listList, err := loader.LoadList(db)
	if err != nil {
		logger.Fatalf("load token info failed, %s", err)
	}
	handlers.SetLists(listList)

	tokenBalanceList, err := loader.LoadTokenBalances(db, rerun, rerun_start)
	if err != nil {
		logger.Fatalf("load token balances failed, %s", err)
	}

	err = handlers.SetTokenBalances(tokenBalanceList)
	if err != nil {
		logger.Fatalf("set token balances failed, %s", err)
	}

	err = loader.GetMaxBlockNumberFromDB(db)
	if err != nil {
		logger.Fatalf("get max block number from db failed %s", err)
	}
	trxs, err := loader.LoadTransactionData(inputfile1)
	if err != nil {
		logger.Fatalf("invalid input, %s", err)
	}

	logs, err := loader.LoadLogData(inputfile2)
	if err != nil {
		logger.Fatalf("invalid input, %s", err)
	}

	records := handlers.MixRecords(trxs, logs)

	err = handlers.ProcessRecords(records)
	if err != nil {
		logger.Fatalf("process error, %s", err)
	}

	tokenBalances := handlers.GetTokenBalances()

	tokens := handlers.GetTokenInfo()
	tokenInfos := loader.ConvertTokensToTokenInfos(tokens)

	asc20s := handlers.GetAsc20()
	tokenActivities := loader.ConvertAsc20sToTokenActivities(asc20s)

	inscriptions := handlers.GetInscriptions()
	logEvents := handlers.GetLogEvents()

	err = tidb.ProcessUpsert(db, inscriptions, logEvents, tokenInfos, tokenActivities, tokenBalances)
	if err != nil {
		logger.Fatalf("process error, %s", err)
	}

	logger.Info("Index successed")
}
