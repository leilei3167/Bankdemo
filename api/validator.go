package api

import (
	"github.com/go-playground/validator/v10"
	"github.com/leilei3167/bank/db/util"
)

//声明一个validator
var validCurrency validator.Func = func(fl validator.FieldLevel) bool {
	//Field返回的是反射值,将其转化为空接口后断言是否为string
	if currency, ok := fl.Field().Interface().(string); ok {
		return util.IsSupportedCurrency(currency)
	}
	return false
}
