package api

import (
	"github.com/gin-gonic/gin"
	db "github.com/leilei3167/bank/db/sqlc"
	"github.com/leilei3167/bank/db/util"
	"github.com/lib/pq"
	"net/http"
	"time"
)

//构建所需的数据,在此可用binding来验证输入的字段
type createUserRequest struct {
	Username string `json:"username" binding:"required,alphanum"` //alphanum代表不能是特殊字符,只能是ascii编码
	Password string `json:"password" binding:"required,min=6" `   //必须字段,最短6位
	Fullname string `json:"full_name" binding:"required"`
	Email    string `json:"email" binding:"required,email"` //email代表必须是正确email格式
}

type ResUser struct {
	Username string `json:"username"`
	//HashedPassword    string    `json:"hashed_password"`
	FullName          string    `json:"full_name"`
	Email             string    `json:"email"`
	PasswordChangedAt time.Time `json:"password_changed_at"`
	CreatedAt         time.Time `json:"created_at"`
}

func (server *Server) createUser(ctx *gin.Context) {
	var req createUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorRespones(err)) //创建一个函数,将err转化为k-v键值对
		return
	}
	//处理密码
	hashedpassword, err := util.HashPassword(req.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorRespones(err))
		return
	}
	arg := db.CreateUserParams{
		Username:       req.Username,
		HashedPassword: hashedpassword,
		FullName:       req.Fullname,
		Email:          req.Email,
	}
	user, err := server.store.CreateUser(ctx, arg)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			//log.Println(pqErr.Code.Name())
			switch pqErr.Code.Name() {
			case "unique_violation":
				ctx.JSON(http.StatusForbidden, errorRespones(err))
			}

		}
		ctx.JSON(http.StatusInternalServerError, errorRespones(err))
		return
	}
	//不应该将hash之后的密码也返回
	resp := ResUser{
		Username:          user.Username,
		FullName:          user.FullName,
		Email:             user.Email,
		PasswordChangedAt: user.PasswordChangedAt,
		CreatedAt:         user.CreatedAt,
	}
	ctx.JSON(http.StatusOK, resp)

}
