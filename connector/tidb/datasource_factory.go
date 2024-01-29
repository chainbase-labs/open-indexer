package tidb

import (
	"encoding/json"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
	"log"
	logger2 "open-indexer/logger"
	"open-indexer/model"
	"os"
	"reflect"
	"sync"
)

var (
	db   *gorm.DB
	once sync.Once
)

var tblCreateSqlMap = make(map[string]string)
var initFilePath = "./data/init/"

func init() {
	tblCreateSqlMap["token_info"] = "CREATE TABLE IF NOT EXISTS `token_info` (\n    `block_timestamp` datetime(3) NOT NULL COMMENT 'Timestamp of the block containing the inscription (matches block_timestamp in transactions table)',\n    `block_number` bigint(20) NOT NULL COMMENT 'Block number containing the inscription (matches block_number in transactions table)',\n    `tx_index` int(11) NOT NULL COMMENT 'Index of the transaction containing the inscription (matches transaction_index in transactions table)',\n    `tx_hash` varchar(66) NOT NULL COMMENT 'Unique identifier of the transaction containing the inscription (matches hash in transactions table)',\n    `tick` varchar(255) NOT NULL COMMENT 'Token tick',\n    `max_supply` decimal(38, 0) DEFAULT NULL COMMENT 'Max supply',\n    `lim` decimal(38, 0) DEFAULT NULL COMMENT 'Limit of each mint',\n    `wlim` decimal(38, 0) DEFAULT NULL COMMENT 'Limit of each address can maximum mint',\n    `dec` int(11) DEFAULT NULL COMMENT 'Decimal for minimum divie',\n    `creator` varchar(42) DEFAULT NULL COMMENT 'Address originating the inscription (matches from_address in transactions table)',\n    `minted` decimal(38, 0) DEFAULT '0',\n    `holders` decimal(38, 0) DEFAULT '0',\n    `txs` decimal(38, 0) DEFAULT '0',\n    `updated_timestamp` timestamp(3) NULL DEFAULT NULL,\n    `completed_timestamp` timestamp(3) NULL DEFAULT NULL,\n    `id` varchar(255) DEFAULT NULL,\n    PRIMARY KEY (`tick`)\n    );\n"
	tblCreateSqlMap["token_activities"] = "CREATE TABLE IF NOT EXISTS `token_activities` (\n                                    `block_timestamp` datetime(3) NOT NULL COMMENT 'Timestamp of the block containing the inscription (matches block_timestamp in transactions table)',\n                                    `block_number` bigint(20) NOT NULL COMMENT 'Block number containing the inscription (matches block_number in transactions table)',\n                                    `tx_index` int(11) NOT NULL COMMENT 'Index of the transaction containing the inscription (matches transaction_index in transactions table)',\n                                    `tx_hash` varchar(66) NOT NULL COMMENT 'Unique identifier of the transaction containing the inscription (matches hash in transactions table)',\n                                    `log_index` int(11) NOT NULL COMMENT 'Index of the log within the transaction',\n                                    `type` varchar(255) NOT NULL COMMENT 'mint  transfer  burn',\n                                    `tick` varchar(255) NOT NULL COMMENT 'Token tick',\n                                    `id` varchar(255) NOT NULL COMMENT 'Unique identifier of the inscription',\n                                    `amt` decimal(38, 0) DEFAULT NULL COMMENT 'Mint amount',\n                                    `from_address` varchar(42) DEFAULT NULL COMMENT 'Address sending the inscription (matches from_address in transactions table)',\n                                    `to_address` varchar(42) DEFAULT NULL COMMENT 'Address receiving the inscription (match to_address in transactions table)',\n                                    PRIMARY KEY (\n                                                 `id`, `log_index`, `tx_index`, `tick`,\n                                                 `type`\n                                        )\n);\n"
	tblCreateSqlMap["token_balances"] = "CREATE TABLE `token_balances` (\n                                  `block_number` bigint(20) unsigned DEFAULT NULL COMMENT 'Block number containing the transaction',\n                                  `block_timestamp` datetime(3) DEFAULT NULL COMMENT 'Block timestamp containing the transaction',\n                                  `tick` varchar(255) NOT NULL COMMENT 'Token tick',\n                                  `wallet_address` varchar(42) NOT NULL COMMENT 'Address of owner',\n                                  `total_supply` decimal(38, 0) DEFAULT NULL COMMENT 'Max supply',\n                                  `amount` decimal(38, 0) DEFAULT NULL COMMENT 'The balance of wallet balance at the corresponding block height',\n                                  PRIMARY KEY (`tick`, `wallet_address`)\n);\n"
	tblCreateSqlMap["token_balances_his"] = "CREATE TABLE `token_balances_his`\n(\n    `block_number`    bigint(20) NOT NULL COMMENT 'Block number containing the transaction',\n    `block_timestamp` datetime(3) DEFAULT NULL COMMENT 'Block timestamp containing the transaction',\n    `tick`            varchar(255) NOT NULL COMMENT 'Token tick',\n    `wallet_address`  varchar(42)  NOT NULL COMMENT 'Address of owner',\n    `total_supply`    decimal(38, 0) DEFAULT NULL COMMENT 'Max supply',\n    `amount`          decimal(38, 0) DEFAULT NULL COMMENT 'The balance of wallet balance at the corresponding block height',\n    PRIMARY KEY (`block_number`,`tick`, `wallet_address`)\n);"
}

type Config struct {
	TiDBUser     string `json:"tidb_user"`
	TiDBPassword string `json:"tidb_password"`
	TiDBHost     string `json:"tidb_host"`
	TiDBPort     string `json:"tidb_port"`
	TiDBDBName   string `json:"tidb_db_name"`
}

func getConfigFromFile(file_name string) (*Config, error) {
	configFile, err := os.Open(file_name)
	if err != nil {
		return nil, err
	}
	defer configFile.Close()

	var config Config
	jsonParser := json.NewDecoder(configFile)
	if err = jsonParser.Decode(&config); err != nil {
		return nil, err
	}
	return &config, nil
}

func createDB(tidb_user string, tidb_password string, tidb_host string, tidb_port string, tidb_db_name string) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True",
		tidb_user, tidb_password, tidb_host, tidb_port, tidb_db_name)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Error),
	})

	if err != nil {
		return nil, err
	}

	return db, nil
}

func JudgeTableExistOrNot(db *gorm.DB, tableName string) (bool, error) {
	var count int
	db.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = ?", tableName).Scan(&count)

	if count == 0 {
		return false, nil
	}
	return true, nil
}

func CreateTableIfNotExist[T any](db *gorm.DB, table T, tableName string) error {
	exist, err := JudgeTableExistOrNot(db, tableName)
	if !exist {
		fileSql, ok := tblCreateSqlMap[tableName]
		if !ok {
			tType := reflect.TypeOf(*new(T))
			instance := reflect.New(tType).Interface()
			err := db.AutoMigrate(instance)
			if err != nil {
				log.Fatalf("Create table %s failed: %v", tableName, err)
				return err
			}
		} else {
			err = db.Exec(fileSql).Error
			if err != nil {
				log.Fatalf("Create table %s failed: %v", tableName, err)
				return err
			}
		}
	}
	if err != nil {
		return err
	}
	return nil
}

func GetDBInstanceByConfigFile(file_name string) (*gorm.DB, error) {
	var err error
	once.Do(func() {
		var config *Config
		config, err = getConfigFromFile("./connector/tidb/" + file_name)
		if err != nil {
			return
		}
		db, err = createDB(config.TiDBUser, config.TiDBPassword, config.TiDBHost, config.TiDBPort, config.TiDBDBName)
	})
	return db, err
}

func GetDBInstanceByEnv() (*gorm.DB, error) {
	var err error
	once.Do(func() {
		db, err = createDB(os.Getenv("tidb_user"), os.Getenv("tidb_password"), os.Getenv("tidb_host"), os.Getenv("tidb_port"), os.Getenv("tidb_db_name"))
	})
	return db, err
}

func batchUpsert[T any](db *gorm.DB, datas []T, batchSize int, table_name string) error {
	if len(datas) == 0 {
		return nil
	}

	var logger = logger2.GetLogger()
	for i := 0; i < len(datas); i += batchSize {
		end := i + batchSize
		if end > len(datas) {
			end = len(datas)
		}

		err := db.Table(table_name).Clauses(clause.OnConflict{
			UpdateAll: true,
		}).Create(datas[i:end]).Error

		if err != nil {
			return err
		}
	}

	logger.Infof("Upsert into db successed, items %d %s", len(datas), table_name)
	return nil
}

func ProcessUpsert(db *gorm.DB, inscriptions []*model.Inscription, logEvents []*model.EvmLog, tokens []*model.TokenInfo, tokenActivities []*model.TokenActivity, tokenBalances map[string]map[string]*model.TokenBalance) error {
	CreateTableIfNotExist(db, model.Inscription{}, model.Inscription{}.TableName())
	CreateTableIfNotExist(db, model.EvmLog{}, model.EvmLog{}.TableName())
	CreateTableIfNotExist(db, model.TokenInfo{}, model.TokenInfo{}.TableName())
	CreateTableIfNotExist(db, model.TokenActivity{}, model.TokenActivity{}.TableName())
	CreateTableIfNotExist(db, model.TokenBalance{}, model.TokenBalance{}.TableName())
	CreateTableIfNotExist(db, model.TokenBalanceHis{}, model.TokenBalanceHis{}.TableName())

	var defaultBatchSize = 200
	tx := db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	if err := batchUpsert(tx, inscriptions, defaultBatchSize, model.Inscription{}.TableName()); err != nil {
		tx.Rollback()
		return err
	}

	if err := batchUpsert(tx, logEvents, defaultBatchSize, model.EvmLog{}.TableName()); err != nil {
		tx.Rollback()
		return err
	}

	if err := batchUpsert(tx, tokens, defaultBatchSize, model.TokenInfo{}.TableName()); err != nil {
		tx.Rollback()
		return err
	}

	if err := batchUpsert(tx, tokenActivities, defaultBatchSize, model.TokenActivity{}.TableName()); err != nil {
		tx.Rollback()
		return err
	}

	var list []*model.TokenBalance
	for _, holders := range tokenBalances {
		for _, holder := range holders {
			list = append(list, holder)
		}
	}

	if err := batchUpsert(tx, list, defaultBatchSize, model.TokenBalance{}.TableName()); err != nil {
		tx.Rollback()
		return err
	}

	if err := batchUpsert(tx, list, defaultBatchSize, model.TokenBalanceHis{}.TableName()); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}
