package pggorm

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Db struct {
	db *gorm.DB
}

func NewDB(connUrl string) (*Db, error) {
	db, err := gorm.Open(postgres.Open(connUrl), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, err
	}

	db = db.Session(&gorm.Session{
		PrepareStmt: false,
	})

	return &Db{db: db}, nil
}

func (db *Db) DB() *gorm.DB {
	return db.db
}

func (db *Db) Close() error {
	sqlDb, err := db.db.DB()
	if err != nil {
		return err
	} else {
		return sqlDb.Close()
	}
}
