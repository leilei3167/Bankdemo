package api

import (
	"github.com/gin-gonic/gin"
	db "github.com/leilei3167/bank/db/sqlc"
)

//因为涉及到数据库的交互,所以嵌入store,router为路由
type Server struct {
	store  db.Store
	router *gin.Engine
}

//链接到数据库之后传入store,返回新的Server实例
func NewServer(store db.Store) *Server {
	server := &Server{store: store}
	router := gin.Default()
	//暂时添加创建的api
	//传入多个处理器的话中间的是中间件
	//处理器函数都围绕server结构体构建,因为其中包括了数据库的交互
	router.POST("/accounts", server.createAccount)
	router.GET("/accounts/:id", server.getAccount) //:id告诉gin id字段是参数
	router.GET("/accounts", server.ListAccount)
	server.router = router
	return server

}

//在指定地址开启服务,外部可访问此方法
func (server *Server) Start(addr string) error {
	return server.router.Run(addr) //后面添加优雅关闭逻辑

}

//将各个地方的错误处理封装成函数,返回键值对,简化代码
func errorRespones(err error) gin.H {
	return gin.H{"error": err.Error()}

}
