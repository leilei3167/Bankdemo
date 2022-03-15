package db

import (
	"database/sql"
	"log"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

//数据库链接设置为常量
const (
	dbDriver = "postgres"
	dbSource = "postgresql://root:8888@localhost:5432/simple_bank?sslmode=disable"
)

var testQueries *Queries
var testDB *sql.DB //全局db 方便其他测试函数使用

//整个包的测试主入口
func TestMain(m *testing.M) {
	var err error
	testDB, err = sql.Open(dbDriver, dbSource)
	if err != nil {
		log.Fatal("无法链接到数据库:", err)
	}
	//拿到conn后创建Queries实例,
	testQueries = New(testDB)
	os.Exit(m.Run())
}
