package common

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"os"
	"time"
)

const (
	MAX_NUM_CONN  = 1000
	MAX_IDLE_CONN = 100
	MAX_LIFETIME  = 3
)

var (
	databaseURL1 = env("JAWSDB_URL", "root:toor@tcp(shard1:3306)/message?parseTime=true")
	databaseURL2 = env("JAWSDB_URL", "root:toor@tcp(shard2:3306)/message?parseTime=true")
	databaseURL3 = env("JAWSDB_URL", "root:toor@tcp(shard3:3306)/message?parseTime=true")
	databaseURL4 = env("JAWSDB_URL", "root:toor@tcp(shard4:3306)/message?parseTime=true")

	//databaseURL1 = env("JAWSDB_URL", "root:toor@tcp(127.0.0.1:3370)/message?parseTime=true")
	//databaseURL2 = env("JAWSDB_URL", "root:toor@tcp(127.0.0.1:3371)/message?parseTime=true")
	//databaseURL3 = env("JAWSDB_URL", "root:toor@tcp(127.0.0.1:3372)/message?parseTime=true")
	//databaseURL4 = env("JAWSDB_URL", "root:toor@tcp(127.0.0.1:3373)/message?parseTime=true")

	Db1 *sql.DB
	Db2 *sql.DB
	Db3 *sql.DB
	Db4 *sql.DB
)

func setDB(pathDB string) *sql.DB {

	db, err1 := sql.Open("mysql", pathDB)
	//db, err_1 := sql.Open("mysql", pathDB)
	if err1 != nil {
		log.Println(" ---------------------  ERROR database opening " + pathDB)
		panic(err1)
	}
	err := db.Ping()
	if err != nil {
		log.Println(" ---------------------  ERROR database PING")
		panic(err)
	}
	db.SetConnMaxLifetime(time.Minute * MAX_LIFETIME)
	db.SetMaxOpenConns(MAX_NUM_CONN)
	db.SetMaxIdleConns(MAX_IDLE_CONN)
	return db
}

func InitDB() {
	//db part
	time.Sleep(2 * time.Second)
	Db1 = setDB(databaseURL1)
	Db2 = setDB(databaseURL2)
	Db3 = setDB(databaseURL3)
	Db4 = setDB(databaseURL4)
}

func env(key, fallbackValue string) string {
	s, ok := os.LookupEnv(key)
	if !ok {
		return fallbackValue
	}
	return s
}
