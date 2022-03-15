package db

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTransferTx(t *testing.T) {
	store := NewStore(testDB)
	//生成2个随机的账户来转账
	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)

	//必须谨慎处理事务,不小心处理并发的话将会是一个噩梦

	//5个协程,并发从账户1转到账户2,每次转账10块
	n := 5
	amount := int64(10)

	errs := make(chan error)               //接收错误
	results := make(chan TransferTxResult) //接收结果

	for i := 0; i < n; i++ {
		go func() {
			result, err := store.TransferTx(context.Background(), TransferTxParams{
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

		ToEntry := result.ToEntry
		require.NotEmpty(t, ToEntry)
		require.Equal(t, account2.ID, ToEntry.AccountID)
		require.Equal(t, amount, ToEntry.Amount)
		require.NotZero(t, ToEntry.CreatedAt)

		_, err = store.GetEntry(context.Background(), ToEntry.ID)

		//TODO: 还有balance的没有做

	}

}
