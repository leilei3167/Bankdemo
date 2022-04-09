package main

import (
	"database/sql"
	"github.com/leilei3167/bank/db/util"
	"log"

	"github.com/leilei3167/bank/api"
	db "github.com/leilei3167/bank/db/sqlc"
	_ "github.com/lib/pq"
)

/*const (
	dbDriver   = "postgres"
	dbSource   = "postgresql://root:123456@localhost:5432/bank?sslmode=disable"
	serverAddr = "0.0.0.0:8080"
)*/

func main() {
	//先连接数据库
	config, err := util.LoadConfig(".") //"."代表当前文件夹
	if err != nil {
		log.Fatal("读取配置文件失败!", err)
	}
	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("无法链接到数据库:", err)
	}
	//构建Server
	store := db.NewStore(conn)
	server := api.NewServer(store)
	err = server.Start(config.ServerAdress)
	if err != nil {
		log.Fatal("无法启动web服务:", err)
	}
}
