package tidb

import (
	"encoding/json"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
	"open-indexer/model"
	"os"
	"sync"
)

var (
	db   *gorm.DB
	once sync.Once
)

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
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, err
	}

	return db, nil
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

func Upsert[T any](db *gorm.DB, datas []T) error {
	for _, data := range datas {
		if err := db.Clauses(clause.OnConflict{UpdateAll: true}).Create(&data).Error; err != nil {
			return err
		}
	}
	return nil
}

func ProcessUpsert(db *gorm.DB, inscriptions []*model.Inscription, logEvents []*model.EvmLog, tokens []*model.TokenInfo, tokenActivities []*model.TokenActivity, tokenBalances map[string]map[string]*model.TokenBalance) error {
	tx := db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	if err := Upsert(tx, inscriptions); err != nil {
		tx.Rollback()
		return err
	}

	if err := Upsert(tx, logEvents); err != nil {
		tx.Rollback()
		return err
	}

	if err := Upsert(tx, tokens); err != nil {
		tx.Rollback()
		return err
	}

	if err := Upsert(tx, tokenActivities); err != nil {
		tx.Rollback()
		return err
	}

	var list []*model.TokenBalance
	for _, holders := range tokenBalances {
		for _, holder := range holders {
			list = append(list, holder)
		}
	}

	if err := Upsert(tx, list); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}
