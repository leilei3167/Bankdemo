package main

import (
	"database/sql"
	"log"

	"github.com/leilei3167/bank/api"
	db "github.com/leilei3167/bank/db/sqlc"
	_ "github.com/lib/pq"
)

const (
	dbDriver   = "postgres"
	dbSource   = "postgresql://root:8888@localhost:5432/simple_bank?sslmode=disable"
	serverAddr = "0.0.0.0:8080"
)

func main() {

	conn, err := sql.Open(dbDriver, dbSource)
	if err != nil {
		log.Fatal("无法链接到数据库:", err)
	}
	store := db.NewStore(conn)
	server := api.NewServer(store)
	err = server.Start(serverAddr)
	if err != nil {
		log.Fatal("无法启动web服务:", err)
	}
}
