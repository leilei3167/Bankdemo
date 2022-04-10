package db

import (
	"context"
	"github.com/leilei3167/bank/db/util"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func createRandomUser(t *testing.T) User {
	//定义要传入的参数
	hashedPassword, err := util.HashPassword(util.RandomString(6))
	require.NoError(t, err)
	arg := CreateUserParams{
		Username:       util.RandOwner(),
		HashedPassword: hashedPassword,
		FullName:       util.RandOwner(),
		Email:          util.RandomEmail(),
	}
	//接收返回结果和错误
	user, err := testQueries.CreateUser(context.Background(), arg)
	//检查是否没有错误
	require.NoError(t, err)
	//检查返回的account是否为空
	require.NotEmpty(t, user)
	//检查结果是否与输入值相等
	require.Equal(t, arg.Username, user.Username)
	require.Equal(t, arg.HashedPassword, user.HashedPassword)
	require.Equal(t, arg.Email, user.Email)

	//检查数据库是否自动生成字段
	require.True(t, user.PasswordChangedAt.IsZero())
	require.NotZero(t, user.CreatedAt)
	return user
}

//测试插入
func TestCreateUser(t *testing.T) {
	createRandomUser(t)
}

//测试Get,注意每个单元测试都要独立,不能依赖其他测试创建的数据,因此都自己创建记录
func TestGetUser(t *testing.T) {
	user1 := createRandomUser(t)

	user2, err := testQueries.GetUser(context.Background(), user1.Username)
	require.NoError(t, err)
	require.NotEmpty(t, user2)
	//要求字段都相等
	require.Equal(t, user1.Username, user2.Username)
	require.Equal(t, user1.HashedPassword, user2.HashedPassword)
	require.Equal(t, user1.FullName, user2.FullName)
	require.Equal(t, user1.Email, user2.Email)
	//对于时间戳
	require.WithinDuration(t, user1.CreatedAt, user2.CreatedAt, time.Second)
	require.WithinDuration(t, user1.PasswordChangedAt, user2.PasswordChangedAt, time.Second)
}
