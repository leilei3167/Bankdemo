package api

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/leilei3167/bank/db/sqlc"
)

//构建所需的数据,在此可用binding来验证输入的字段
type createAccountRequest struct {
	Owner    string `json:"owner" binding:"required"`
	Currency string `json:"currency" binding:"required,currency" ` //必须字段
} //只允许传入owner 和 币种,余额创建时默认为0

func (server *Server) createAccount(ctx *gin.Context) {
	var req createAccountRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorRespones(err)) //创建一个函数,将err转化为k-v键值对
		return
	}
	//没有错误的话执行创建,此时req已经被填充了字段
	arg := db.CreateAccountParams{
		Owner:    req.Owner,
		Currency: req.Currency,
		Balance:  0, //金额初始化为0
	}
	account, err := server.store.CreateAccount(ctx, arg)
	if err != nil {
		//错误返回给前端
		ctx.JSON(http.StatusInternalServerError, errorRespones(err))
		return
	}
	//没有错误即处理完成,返回成功消息和account
	ctx.JSON(http.StatusOK, account)

}

//uri无法像json一样从正文获取
type getAccountRequest struct {
	//uri标签告诉gin,参数的名称
	ID int64 `uri:"id" binding:"required,min=1" ` //非空,并且最小值为1
}

func (server *Server) getAccount(ctx *gin.Context) {
	var req getAccountRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorRespones(err)) //创建一个函数,将err转化为k-v键值对
		return
	}
	account, err := server.store.GetAccount(ctx, req.ID)
	if err != nil { //此处错误有2种,一种查不到,一种是查询出错
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorRespones(err))
			return
		}
		//其他错误
		ctx.JSON(http.StatusInternalServerError, errorRespones(err))

	}
	//没有错误
	ctx.JSON(http.StatusOK, account)

}

//分页显示数据
type ListAccountRequest struct {
	//uri标签告诉gin,参数的名称
	PageID   int32 `form:"page_id" binding:"required,min=1" `          //非空,并且最小值为1
	PageSize int32 `form:"page_size" binding:"required,min=5,max=10" ` //非空,并且最小值为1
}

func (server *Server) ListAccount(ctx *gin.Context) {
	var req ListAccountRequest
	//将form绑定到req
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorRespones(err)) //创建一个函数,将err转化为k-v键值对
		return
	}

	arg := db.ListAccountsParams{
		Limit:  req.PageSize,                    //页面的大小,5-10
		Offset: (req.PageID - 1) * req.PageSize, //第几页
	}

	account, err := server.store.ListAccounts(ctx, arg)
	if err != nil { //此处错误有2种,一种查不到,一种是查询出错
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorRespones(err))
			return
		}
		//其他错误
		ctx.JSON(http.StatusInternalServerError, errorRespones(err))

	}
	//没有错误
	ctx.JSON(http.StatusOK, account)

}
