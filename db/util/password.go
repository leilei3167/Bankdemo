package util

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

//将裸密码哈希
func HashPassword(password string) (string, error) {
	hashedpassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost) //默认cost=10
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %v", err)
	}
	return string(hashedpassword), nil
}

//检查裸密码是否和哈希的密码一致
func CheckPassword(password string, hashedPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))

}
