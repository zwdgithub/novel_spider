package db

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	logger "novel_spider/log"
	"time"
)

type MysqlConfig struct {
	User         string
	Pass         string
	DBName       string
	Addr         string        // for trace
	Port         string        // port
	Active       int           // pool
	Idle         int           // pool
	IdleTimeout  time.Duration // connect max life time.
	QueryTimeout time.Duration // query sql timeout
	ExecTimeout  time.Duration // execute sql timeout
	TranTimeout  time.Duration // transaction sql timeout
	Encoding     string
}

func LoadMysqlConfig(filePath string) *MysqlConfig {
	var mysqlConf MysqlConfig
	file, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatal(err)
	}
	err = yaml.Unmarshal(file, &mysqlConf)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(mysqlConf)
	return &mysqlConf
}

func New(conf *MysqlConfig) *gorm.DB {
	initStr := "%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=true&loc=Local"
	fmt.Println(fmt.Sprintf(initStr, conf.User, conf.Pass, conf.Addr, conf.Port, conf.DBName, conf.Encoding))
	db, err := gorm.Open("mysql", fmt.Sprintf(initStr, conf.User, conf.Pass, conf.Addr, conf.Port, conf.DBName, conf.Encoding))
	if err != nil {
		log.Fatal("mysql connection error ", err)
	}
	db.LogMode(false)
	db.SingularTable(true)
	// 空闲时最大连接数
	db.DB().SetMaxIdleConns(10)
	// 最大连接数
	db.DB().SetMaxOpenConns(100)
	db.DB().SetConnMaxLifetime(60 * time.Second)
	db.SetLogger(logger.DBLogger{})
	return db
}
