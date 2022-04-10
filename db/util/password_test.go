package util

import (
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"testing"
)

func TestHashPassword(t *testing.T) {
	password := RandomString(6)

	hashedpassword1, err := HashPassword(password)
	require.NoError(t, err)
	require.NotEmpty(t, hashedpassword1)

	err = CheckPassword(password, hashedpassword1)
	require.NoError(t, err)

	//错误示范
	wrongPassword := RandomString(7)
	err = CheckPassword(wrongPassword, hashedpassword1)
	require.EqualError(t, err, bcrypt.ErrMismatchedHashAndPassword.Error())

	//两次哈希得到的密码应该不一致(底层用了随机salt,因此同一密码多次哈希值也不同)
	hashedpassword2, err := HashPassword(password)
	require.NoError(t, err)
	require.NotEmpty(t, hashedpassword2)
	require.NotEqual(t, hashedpassword1, hashedpassword2)

}
