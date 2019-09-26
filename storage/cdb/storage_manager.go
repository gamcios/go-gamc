// Copyright (C) 2018 go-gamc authors
//
// This file is part of the go-gamc library.
//
// the go-gamc library is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// the go-gamc library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with the go-gamc library.  If not, see <http://www.gnu.org/licenses/>.
//

package cdb

import (
	"gamc.pro/gamcio/go-gamc/conf"
	"gamc.pro/gamcio/go-gamc/util/config"
	"fmt"

	"gamc.pro/gamcio/go-gamc/util/logging"
	"errors"
	"github.com/sirupsen/logrus"
)

const (
	TypeLevelDB = "levelDB"
)

type DbConfig struct {
	DbType      string `yaml:"db_type"`
	EnableBatch bool   `yaml:"enable_batch"`
	DbDir       string `yaml:"db_dir"`
}

func GetDbConfig(config *config.Config) *DbConfig {
	dbConf := new(DbConfig)
	config.GetObject("database", dbConf)
	if dbConf != nil {
		if dbConf.DbType == "" {
			dbConf.DbType = TypeLevelDB
		}
	} else {
		dbConf = NewDefaultDbConfig()
	}
	if dbConf.DbDir == "" {
		chaincfg := conf.GetChainConfig(config)
		dbPath := chaincfg.Datadir + "\\chain"
		dbConf.DbDir = dbPath
	}
	return dbConf
}

func NewDefaultDbConfig() *DbConfig {
	return &DbConfig{
		TypeLevelDB,
		false,
		"",
	}
}

func NewDB(config *config.Config) (Storage, error) {
	dbcfg := GetDbConfig(config)
	if dbcfg.DbType == TypeLevelDB {
		db, err := NewLevelDB(dbcfg, 16, 500)
		if err != nil {
			logging.CLog().WithFields(logrus.Fields{
				"dir": dbcfg.DbDir,
				"err": err,
			}).Error("Failed to new a levelDB instance.")
			return nil, err
		}
		return db, nil
	} else {
		return nil, errors.New(fmt.Sprintf("Does not support the %s database.", dbcfg.DbType))
	}

}

type Database struct {
	Db Storage
}

func (d *Database) Has(key []byte) (bool, error) {
	return d.Db.Has(key)
}
func (d *Database) Get(key []byte) ([]byte, error) {
	return d.Db.Get(key)
}
func (d *Database) Put(key, value []byte) error {
	return d.Db.Put(key, value)
}
func (d *Database) Delete(key []byte) error {
	return d.Db.Delete(key)
}
func (d *Database) Close() error {
	return d.Db.Close()
}

func (d *Database) ValueSize() int {
	return 0
}

func (d *Database) EnableBatch() {

}

func (d *Database) DisableBatch() {

}

func (d *Database) Flush() error {
	return nil
}
