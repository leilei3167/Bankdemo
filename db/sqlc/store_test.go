package db

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTransferTx(t *testing.T) {
	store := NewStore(testDB)
	//生成2个随机的账户来转账
	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)
	fmt.Println(">>>转账之前:", account1.Balance, account2.Balance)
	//必须谨慎处理事务,不小心处理并发的话将会是一个噩梦

	//5个协程,并发从账户1转到账户2,每次转账10块
	n := 5
	amount := int64(10)

	errs := make(chan error)               //接收错误
	results := make(chan TransferTxResult) //接收结果

	for i := 0; i < n; i++ {
		//用于标记协程
		txName := fmt.Sprintf("tx %d", i+1)
		go func() {
			ctx := context.WithValue(context.Background(), txKey, txName)
			result, err := store.TransferTx(ctx, TransferTxParams{
				FromAccountID: account1.ID,
				ToAccountID:   account2.ID,
				Amount:        amount, //每次转账10块
			})
			//将错误发送到主协程
			errs <- err
			results <- result
		}()

	}

	//在外部进行检查
	existed := make(map[int]bool)
	for i := 0; i < n; i++ {
		err := <-errs
		require.NoError(t, err)

		result := <-results
		//结果不能为空
		require.NotEmpty(t, result)

		//检查转账结果
		transfer := result.Transfer
		require.NotEmpty(t, transfer)
		require.Equal(t, account1.ID, transfer.FromAccountID) //结果里的From账户应该该account1的id相等
		require.Equal(t, account2.ID, transfer.ToAccountID)
		require.Equal(t, amount, transfer.Amount)
		require.NotZero(t, transfer.ID)
		require.NotZero(t, transfer.CreatedAt)

		//应该能查询到转账记录
		_, err = store.GetTransfer(context.Background(), transfer.ID)
		require.NoError(t, err)

		//检查entry结果
		fromEntry := result.FromEntry
		require.NotEmpty(t, fromEntry)
		require.Equal(t, account1.ID, fromEntry.AccountID)
		require.Equal(t, -amount, fromEntry.Amount)
		require.NotZero(t, fromEntry.CreatedAt)

		_, err = store.GetEntry(context.Background(), fromEntry.ID)
		require.NoError(t, err)

		ToEntry := result.ToEntry
		require.NotEmpty(t, ToEntry)
		require.Equal(t, account2.ID, ToEntry.AccountID)
		require.Equal(t, amount, ToEntry.Amount)
		require.NotZero(t, ToEntry.CreatedAt)

		_, err = store.GetEntry(context.Background(), ToEntry.ID)
		require.NoError(t, err)

		fromAccount := result.FromAccount
		require.NotEmpty(t, fromAccount)
		require.Equal(t, account1.ID, fromAccount.ID)

		ToAccount := result.ToAccount
		require.NotEmpty(t, ToAccount)
		require.Equal(t, account2.ID, ToAccount.ID)

		//检查金额
		fmt.Println(">>>转账中:", fromAccount.Balance, ToAccount.Balance)

		diff1 := account1.Balance - fromAccount.Balance //账户1转出的金额
		diff2 := ToAccount.Balance - account2.Balance   //收到的金额
		require.Equal(t, diff1, diff2)
		require.True(t, diff1 > 0) //差值必须为正

		//这个差值可以被每笔交易的金额整除,每转账1次,就会减少1被amount的金额
		require.True(t, diff1%amount == 0)

		k := int(diff1 / amount)
		require.True(t, k >= 1 && k <= n)

		require.NotContains(t, existed, k) //k必须是唯一的

	}
	//转账结束后检查两个账户的余额
	upadateAccount1, err := testQueries.GetAccount(context.Background(), account1.ID)
	require.NoError(t, err)
	require.Equal(t, account1.Balance-int64(n)*amount, upadateAccount1.Balance)

	upadateAccount2, err := testQueries.GetAccount(context.Background(), account2.ID)
	fmt.Println(">>>所有转账之后:", upadateAccount1.Balance, upadateAccount2.Balance)

	require.NoError(t, err)
	require.Equal(t, account2.Balance+int64(n)*amount, upadateAccount2.Balance)

}
