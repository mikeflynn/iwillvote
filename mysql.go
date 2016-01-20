package main

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

var myConfig *MySQLConfig

func NewMySQL() *MySQLConfig {
	if myConfig == nil {
		dbHostname := os.Getenv("MYSQL_HOSTNAME")
		dbUsername := os.Getenv("MYSQL_USERNAME")
		dbPassword := os.Getenv("MYSQL_PASSWORD")
		dbDatabase := os.Getenv("MYSQL_DATABASE")

		if dbHostname == "" || dbUsername == "" || dbPassword == "" {
			log.Println("MySQL configuration environmental variables missing!")
			os.Exit(1)
		}

		myConfig = &MySQLConfig{
			Host:     dbHostname,
			User:     dbUsername,
			Password: dbPassword,
			Database: dbDatabase,
		}
	}

	return myConfig
}

func SQLNullIfEmpty(v string) sql.NullString {
	vIsNotNull := false
	if v != "" {
		vIsNotNull = true
	}

	return sql.NullString{v, vIsNotNull}
}

type MySQLConfig struct {
	Host       string
	User       string
	Password   string
	Database   string
	Connection *sql.DB
}

type DBRecord interface {
	Save() error
	Load() error
}

func (this *MySQLConfig) Connect() bool {
	if this.Connection == nil {
		conn, err := sql.Open("mysql", this.User+":"+this.Password+"@tcp("+this.Host+":3306)/"+this.Database)
		if err != nil {
			log.Println("Connection to MySQL could not be made.")
			os.Exit(1)
		}

		this.Connection = conn
	}

	return true
}

func (this *MySQLConfig) Select(query string, params ...interface{}) (*sql.Rows, error) {
	this.Connect()

	return this.Connection.Query(query, params...)
}

func (this *MySQLConfig) Insert(query string, params ...interface{}) (int64, error) {
	this.Connect()

	var theErr error

	if stmt, err := this.Connection.Prepare(query); err == nil {
		if res, err := stmt.Exec(params...); err == nil {
			if id, err := res.LastInsertId(); err == nil {
				return id, nil
			} else {
				theErr = err
			}
		} else {
			theErr = err
		}
	} else {
		theErr = err
	}

	return 0, theErr
}

func (this *MySQLConfig) Update(query string, params ...interface{}) (bool, error) {
	this.Connect()

	var theErr error

	if stmt, err := this.Connection.Prepare(query); err == nil {
		if res, err := stmt.Exec(params...); err == nil {
			if affect, err := res.RowsAffected(); err == nil {
				if affect > 0 {
					return true, nil
				} else {
					return false, nil
				}
			} else {
				theErr = err
			}
		} else {
			theErr = err
		}
	} else {
		theErr = err
	}

	return false, theErr
}
