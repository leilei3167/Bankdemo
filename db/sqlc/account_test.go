package db

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/leilei3167/bank/db/util"
	"github.com/stretchr/testify/require"
)

//不会被测试执行
func createRandomAccount(t *testing.T) Account {
	//定义要传入的参数
	user := createRandomUser(t)
	arg := CreateAccountParams{
		Owner:    user.Username,
		Balance:  util.RandomMoney(),
		Currency: util.RandomCurrency(),
	}
	//接收返回结果和错误
	account, err := testQueries.CreateAccount(context.Background(), arg)
	//检查是否没有错误
	require.NoError(t, err)
	//检查返回的account是否为空
	require.NotEmpty(t, account)
	//检查结果是否与输入值相等
	require.Equal(t, arg.Owner, account.Owner)
	require.Equal(t, arg.Balance, account.Balance)
	require.Equal(t, arg.Currency, account.Currency)

	//检查数据库是否自动生成字段
	require.NotZero(t, account.ID)
	require.NotZero(t, account.CreatedAt)
	return account
}

//测试插入
func TestCreateAccount(t *testing.T) {
	createRandomAccount(t)
}

//测试Get,注意每个单元测试都要独立,不能依赖其他测试创建的数据,因此都自己创建记录
func TestGetAccount(t *testing.T) {
	account1 := createRandomAccount(t)

	account2, err := testQueries.GetAccount(context.Background(), account1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, account2)
	//要求字段都相等
	require.Equal(t, account1.ID, account2.ID)
	require.Equal(t, account1.Owner, account2.Owner)
	require.Equal(t, account1.Balance, account2.Balance)
	require.Equal(t, account1.Currency, account2.Currency)
	//对于时间戳
	require.WithinDuration(t, account1.CreatedAt, account2.CreatedAt, time.Second)
}

//测试更新
func TestUpdateAccounts(t *testing.T) {
	account1 := createRandomAccount(t)
	arg := UpadateAccountParams{
		ID:      account1.ID,
		Balance: util.RandomMoney(),
	}

	account2, err := testQueries.UpadateAccount(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, account2)

	require.Equal(t, account1.ID, account2.ID)
	require.Equal(t, account1.Owner, account2.Owner)
	//x修改的值和修改后值应该相等
	require.Equal(t, arg.Balance, account2.Balance)
	require.Equal(t, account1.Currency, account2.Currency)
	//对于时间戳
	require.WithinDuration(t, account1.CreatedAt, account2.CreatedAt, time.Second)

}

//测试删除
func TestDeleteAccount(t *testing.T) {
	account1 := createRandomAccount(t)
	err := testQueries.DeleteAccounts(context.Background(), account1.ID)

	require.NoError(t, err)
	//在执行查询就希望报错
	account2, err := testQueries.GetAccount(context.Background(), account1.ID)
	require.Error(t, err)
	//希望错误是查询不到行的错误
	require.EqualError(t, err, sql.ErrNoRows.Error())
	require.Empty(t, account2)
}

//测试获取列表
func TestListAccounts(t *testing.T) {
	//创建十个用于测试
	for i := 0; i < 10; i++ {
		createRandomAccount(t)
	}
	arg := ListAccountsParams{

		Limit:  5, //只显示5个
		Offset: 5, //跳过前面0个
	}
	accounts, err := testQueries.ListAccounts(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, accounts)

	for _, account := range accounts {

		require.NotEmpty(t, account)

	}

}
