package db

/*给数据库接口测试提供入口*/
import (
	"database/sql"
	"github.com/leilei3167/bank/db/util"
	"log"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

//数据库链接设置为常量

var testQueries *Queries
var testDB *sql.DB //全局db 方便其他测试函数使用

//整个包的测试主入口
func TestMain(m *testing.M) {
	config, err := util.LoadConfig("../..")
	if err != nil {
		log.Fatal("加载配置文件出错", err)
	}
	testDB, err = sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("无法链接到数据库:", err)
	}
	//拿到conn后创建Queries实例,
	testQueries = New(testDB)
	os.Exit(m.Run())
}
