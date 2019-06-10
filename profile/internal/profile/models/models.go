package models

import (
	"github.com/jinzhu/gorm"
	"github.com/lnsp/microlog/common/logger"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var log = logger.New()

type Profile struct {
	gorm.Model
	UserID      uint
	DisplayName string
	ImageURL    string
	Biography   string
}

type DB struct {
	db *gorm.DB
}

func (d *DB) Ping() error {
	if err := d.db.Exec("select 1").Error; err != nil {
		return errors.Wrap(err, "ping failed")
	}
	return nil
}

func (d *DB) Close() error {
	return d.db.Close()
}

func Open(path string) (*DB, error) {
	log.WithFields(logrus.Fields{
		"path": path,
		"type": "postgres",
	}).Info("accessing database")
	db, err := gorm.Open("postgres", path)
	db.SetLogger(log)
	if err != nil {
		return nil, errors.Wrap(err, "could not connect to data source")
	}
	db.AutoMigrate(&Profile{})
	return &DB{db}, nil
}
