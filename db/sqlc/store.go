package db

import (
	"context"
	"database/sql"
	"fmt"
)

//通过内嵌Queries继承其所有方法,并添加更多方法来支持事务
type SQLStore struct {
	*Queries //只适用于单次查询
	db       *sql.DB
}

//定义一个接口用于mock,包含之前数据库交互的所有方法
type Store interface {
	Querier
	TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error)
}

//使用db构建Store实例,
func NewStore(db *sql.DB) *SQLStore {
	return &SQLStore{
		db:      db,
		Queries: New(db), //直接用db New一个Queries
	}
}

//构建事务操作,传入一个ctx和一个回调函数
//不希望其他包调用此函数
func (store *SQLStore) execTx(ctx context.Context, fn func(*Queries) error) error {
	//开始事务,可设置隔离级别(默认读已提交)
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	//此处创建的tx,同时也是实现了DBTX的(方法都满足),调用new生成Queries,用于在事务中组合各种单次数据库操作
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
	FromAccountID int64 `json:"from_account_id"`
	ToAccountID   int64 `json:"to_account_id"`
	Amount        int64 `json:"amount"`
}

//转账的结果,要求转账的记录的表,转出和接收方的账户表,转出和接收的记录表
type TransferTxResult struct {
	Transfer    Transfer `json:"transfer"`
	FromAccount Account  `json:"from_account"`
	ToAccount   Account  `json:"to_account"`
	FromEntry   Entry    `json:"from_entry"`
	ToEntry     Entry    `json:"to_entry"`
}

var txKey = struct{}{}

//写一个转账的处理,需要传入转账的参数结构体,返回结果结构体
//创建转账记录,添加account entries,更新account的balance

func (store *SQLStore) TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error) {
	//创建一个空的结果
	var result TransferTxResult
	//调用
	err := store.execTx(ctx, func(q *Queries) error {
		//fn之内为多个语句的组合,任意一个失败都返回err到execTx,并且回滚

		//1.用Queries调用创建转账记录的方法,并将结果写入result transfer字段
		var err error
		txName := ctx.Value(txKey)
		fmt.Println(txName, "创建Transfer")
		//引用外部函数变量result,构成闭包
		result.Transfer, err = q.CreateTransfer(ctx, CreateTransferParams{
			FromAccountID: arg.FromAccountID,
			ToAccountID:   arg.ToAccountID,
			Amount:        arg.Amount,
		})
		if err != nil {
			return err
		}
		//2.转出转入的双方的Account表
		fmt.Println(txName, "创建Entry1")
		result.FromEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.FromAccountID,
			Amount:    -arg.Amount, //对于转出账户来讲,是负数

		})
		if err != nil {
			return err
		}
		fmt.Println(txName, "创建Entry2")
		result.ToEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.ToAccountID,
			Amount:    arg.Amount, //收入账户为正
		})
		if err != nil {
			return err
		}
		//--------------------转账较为复杂,要考虑死锁问题----------------------
		if arg.FromAccountID < arg.ToAccountID { //确保多个事务都按照相同的顺序执行修改
			//先修改转出方
			result.FromAccount, result.ToAccount, err = addMoney(ctx, q, arg.FromAccountID, -arg.Amount,
				arg.ToAccountID, arg.Amount)

		} else { //否则就先修改转入方
			result.ToAccount, result.FromAccount, err = addMoney(ctx, q, arg.ToAccountID, arg.Amount,
				arg.FromAccountID, -arg.Amount)

		}
		//暂时return nil
		return nil
	})

	return result, err
}

//为精简代码而将其封装为函数
func addMoney(
	ctx context.Context,
	q *Queries,

	accountID1 int64,
	amount1 int64, //应该添加到第一个账户的金额
	accountID2 int64,
	amount2 int64, //应该添加到第2个账户的金额
	//返回修改后的两个账户和一个潜在的错误
) (account1 Account, account2 Account, err error) {

	account1, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
		ID:     accountID1,
		Amount: amount1,
	})
	if err != nil {
		return
	}

	account2, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
		ID:     accountID2,
		Amount: amount2,
	})
	if err != nil {
		return
	}

	return

}
