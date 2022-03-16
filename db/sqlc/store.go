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

func (store *Store) TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error) {
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
		//3.更新balance,要考虑加锁和防止死锁
		//一般来讲,应在Account表中查询到需要修改的账户id,然后对其balance字段进行修改,然后如果不加锁,将会出现问题
		//先查询获取余额
		/* 	fmt.Println(txName,"查询account1")
		account1, err := q.GetAccountForUpdate(ctx, arg.FromAccountID)
		if err != nil {
			return err
		} */
		//再修改
		fmt.Println(txName, "修改account1的余额")
		result.FromAccount, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
			ID:     arg.FromAccountID,
			Amount: -arg.Amount,
		})
		if err != nil {
			return err
		}
		/* 	fmt.Println(txName,"查询account2")
		account2, err := q.GetAccountForUpdate(ctx, arg.ToAccountID)
		if err != nil {
			return err
		} */
		//再修改
		fmt.Println(txName, "修改account2的余额")
		result.ToAccount, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
			ID:     arg.ToAccountID,
			Amount: arg.Amount,
		})
		if err != nil {
			return err
		}

		//暂时return nil
		return nil
	})

	return result, err
}
