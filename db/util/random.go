package util

import (
	"fmt"
	"math/rand"
	"time"
)

//随机种子
func init() {
	rand.Seed(time.Now().UnixNano())

}

//随机整数
func RandomInt(min, max int64) int64 {
	//返回一个介于最大值和最小值之间的随机数
	return min + rand.Int63n(max-min+1) //这一步指的是返回[0,到max-min+1)的随机数,加上min就是之间的了

}

const alp = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

//随机字符串
func RandomString(n int) string {
	//构建字符串生成器
	/*	var sb strings.Builder
		k := len(alp)

		for i := 0; i < n; i++ {
			c := alp[rand.Intn(k)]
			//将随机数对应的字母写入到sb中
			sb.WriteByte(c)
		}
		return sb.String()*/
	b := make([]byte, n)
	for i := range b {
		b[i] = alp[rand.Int63()%int64(len(alp))]
	}

	return string(b)

}

//生成随机的owner 名字
func RandOwner() string {
	//返回6位的随机字符串
	return RandomString(6)

}

//随机钱,0--1000
func RandomMoney() int64 {
	return RandomInt(0, 1000)
}

//随机币种
func RandomCurrency() string {
	currency := []string{RMB, USD, EUR}
	n := len(currency)
	//从随机索引中获取
	return currency[rand.Intn(n)]
}

func RandomEmail() string {
	return fmt.Sprintf("%s@qq.com", RandomString(6))
}
