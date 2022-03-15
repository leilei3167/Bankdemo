package db

import (
	"context"
	"database/sql"
	"fmt"
)

//通过内嵌Queries继承其所有方法,并添加更多方法来支持事务
type Store struct {
	*Queries
	db *sql.DB
}

//使用db构建Store实例,
func NewStore(db *sql.DB) *Store {
	return &Store{
		db:      db,
		Queries: New(db), //直接用db New一个Queries
	}
}

//构建事务操作,传入一个ctx和一个回调函数
//不希望其他包调用此函数
func (store *Store) execTx(ctx context.Context, fn func(*Queries) error) error {
	//开始事务,可设置隔离级别(默认读已提交)
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	//此处创建的tx,同时也是实现了DBTX的(方法都满足),调用new生成Queries
	q := New(tx)
	//现在有了能够再事务中执行操作的Queries
	//执行事务
	err = fn(q)
	//如果出错则必须执行回滚
	if err != nil {
		//要同时处理执行事务的错误和回滚出现的错误
		rbErr := tx.Rollback()
		if rbErr != nil {
			return fmt.Errorf("事务错误:%v,回滚错误:%v", err, rbErr)
		}
		//回滚成功只返回事务失败的错误
		return err
	}
	//如果事务执行成功,则提交,返回的错误直接返回给调用处
	return tx.Commit()

}

//z转账相关的结构体
type TransferTxParams struct {
	FromAccountID sql.NullInt64 `json:"from_account_id"`
	ToAccountID   sql.NullInt64 `json:"to_account_id"`
	Amount        int64         `json:"amount"`
}

//转账的结果,要求转账的记录的表,转出和接收方的账户表,转出和接收的记录表
type TransferTxResult struct {
	Transfer    Transfer `json:"transfer"`
	FromAccount Account  `json:"from_account"`
	ToAccount   Account  `json:"to_account"`
	FromEntry   Entry    `json:"from_entry"`
	ToEntry     Entry    `json:"to_entry"`
}

//写一个转账的处理,需要传入转账的参数结构体,返回结果结构体
//创建转账记录,添加account entries,更新account的balance

func (store *Store) TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error) {
	//创建一个空的结果
	var result TransferTxResult

	//调用
	err := store.execTx(ctx, func(q *Queries) error {
		//用Queries调用创建转账记录的方法,并将结果写入result transfer字段
		var err error
		result.Transfer, err = q.CreateTransfer(ctx, CreateTransferParams{
			FromAccountID: arg.FromAccountID,
			ToAccountID:   arg.ToAccountID,
			Amount:        arg.Amount,
		})
		if err != nil {
			return err
		}
		//暂时return nil
		return nil
	})

	return result, err
}
