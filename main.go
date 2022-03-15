package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	db "github.com/leilei3167/bank/db/sqlc"
	_ "github.com/lib/pq"
)

const (
	dbDriver = "postgres"
	dbSource = "postgresql://root:8888@localhost:5432/simple_bank?sslmode=disable"
)

func main() {
	ctx := context.Background()
	conn, err := sql.Open(dbDriver, dbSource)
	if err != nil {
		log.Fatal("无法链接到数据库:", err)
	}
	//拿到conn后创建Queries实例,
	newQueries := db.New(conn)

	m := db.CreateAccountParams{
		Owner:    "leilei",
		Balance:  10000,
		Currency: "RMB",
	}
	res, err := newQueries.CreateAccount(ctx, m)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(res)
	m1 := db.ListAccountsParams{
		Owner:  "Tom",
		Limit:  10,
		Offset: 0,
	}
	s, err := newQueries.ListAccounts(ctx, m1)
	if err != nil {
		log.Fatal(err)

	}
	fmt.Println(s)
	err = newQueries.DeleteAccounts(ctx, 130)
	if err != nil {
		log.Fatal(err)
	}
	u := db.UpadateAccountParams{
		ID:      129,
		Balance: 1222212121,
	}
	newQueries.UpadateAccount(ctx, u)

	s1, err := newQueries.GetAccount(ctx, 129)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(s1)

}
