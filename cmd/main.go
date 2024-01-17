package main

import (
	"flag"
	"log"
	"open-indexer/connector/tidb"
	"open-indexer/handlers"
	"open-indexer/loader"
)

var (
	inputfile1 string
	inputfile2 string
	outputfile string
)

func init() {
	flag.StringVar(&inputfile1, "transactions", "", "the filename of input data, default(./data/transactions.input.txt)")
	flag.StringVar(&inputfile2, "logs", "", "the filename of input data, default(./data/logs.input.txt)")
	flag.StringVar(&outputfile, "output", "./data/asc20.output.txt", "the filename of output result, default(./data/asc20.output.txt)")

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

	tokenBalanceList, err := loader.LoadTokenBalances(db)
	if err != nil {
		logger.Fatalf("load token balances failed, %s", err)
	}

	err = handlers.SetTokenBalances(tokenBalanceList)
	if err != nil {
		logger.Fatalf("set token balances failed, %s", err)
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

	logger.Info("successed")

	// print
	tokens, userBalances, tokenHolders := handlers.GetInfo()
	loader.DumpTickerInfoMap(outputfile, tokens, userBalances, tokenHolders)

	tokenBalances := handlers.GetTokenBalances()

	tokens = handlers.GetTokenInfo()
	tokenInfos := loader.ConvertTokensToTokenInfos(tokens)

	asc20s := handlers.GetAsc20()
	tokenActivities := loader.ConvertAsc20sToTokenActivities(asc20s)

	inscriptions := handlers.GetInscriptions()
	logEvents := handlers.GetLogEvents()

	tidb.ProcessUpsert(db, inscriptions, logEvents, tokenInfos, tokenActivities, tokenBalances)

}
