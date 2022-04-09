package api

import (
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	db "github.com/leilei3167/bank/db/sqlc"
	"net/http"
)

//定义发起转账所需要的参数
type transferRequest struct {
	FromAccoutID int64  `json:"fromAccoutID,omitempty" binding:"required,min=1"`
	ToAccountID  int64  `json:"toAccountID,omitempty"  binding:"required,min=1"`
	Amout        int64  `json:"amout,omitempty" binding:"required,min=0"`
	Currency     string `json:"currency,omitempty" binding:"required,currency"`
}

func (server *Server) createTransfer(ctx *gin.Context) {
	var req transferRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorRespones(err))
		return
	}
	//构建数据库中创建转账的必须字段
	arg := db.TransferTxParams{
		FromAccountID: req.FromAccoutID,
		ToAccountID:   req.ToAccountID,
		Amount:        req.Amout,
	}
	//需要考虑用户转账的货币种类和自己的账户是否相符
	if !server.validAccount(ctx, req.FromAccoutID, req.Currency) {
		return
	}
	if !server.validAccount(ctx, req.ToAccountID, req.Currency) {
		return
	}

	result, err := server.store.TransferTx(ctx, arg) //gin中的context是实现了context.Context的
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorRespones(err))
		return
	}

	ctx.JSON(http.StatusOK, result)
}

func (server *Server) validAccount(ctx *gin.Context, accountID int64, currency string) bool {
	account, err := server.store.GetAccount(ctx, accountID)
	if err != nil {
		//两种错误
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorRespones(err))
			return false
		}
		ctx.JSON(http.StatusInternalServerError, errorRespones(err))
		return false
	}
	if account.Currency != currency {
		err := fmt.Errorf("account [%v] currency mismatch:[%v]->[%v]",
			accountID, account.Currency, currency)
		ctx.JSON(http.StatusBadRequest, errorRespones(err))
		return false
	}
	return true
}
