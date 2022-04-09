package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	mockdb "github.com/leilei3167/bank/db/mock"
	db "github.com/leilei3167/bank/db/sqlc"
	"github.com/leilei3167/bank/db/util"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func randomAccount() db.Account {
	return db.Account{
		ID:       util.RandomInt(1, 1000),
		Owner:    util.RandOwner(),
		Balance:  util.RandomMoney(),
		Currency: util.RandomCurrency(),
	}

}

func TestGetAccount(t *testing.T) {
	//创建一个随机的账户
	account := randomAccount()

	testCases := []struct {
		name          string
		accountID     int64
		buildStubs    func(store *mockdb.MockStore)                           //每一个测试用例都需要构建独立的mock
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder) //用于检测得到的数据

	}{
		{
			name:      "OKcase",
			accountID: account.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(),
					gomock.Eq(account.ID)).Times(1).Return(account, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchAccount(t, recorder.Body, account)
			},
		},

		{
			name:      "NotFound",
			accountID: account.ID, //使用相同的ID也可以,因为每个case模拟的store都是独立的
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(),
					//期望的到查询不到的错误和返回一个空的Account结构体
					gomock.Eq(account.ID)).Times(1).
					Return(db.Account{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code) //code应该和api错误时一致
				//requireBodyMatchAccount(t, recorder.Body, account) 不需要对比
			},
		},

		{
			name:      "InternalError",
			accountID: account.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(),
					gomock.Eq(account.ID)).Times(1).
					Return(db.Account{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code) //code应该和api错误时一致
				//requireBodyMatchAccount(t, recorder.Body, account) 不需要对比
			},
		},
		{
			name:      "BadRequest",
			accountID: 0, //输入一个无效的id
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(),
					gomock.Any()).Times(0)
				//因为无效的ID不会到数据库查询 所以将其去除
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code) //code应该和api错误时一致
				//requireBodyMatchAccount(t, recorder.Body, account) 不需要对比
			},
		},
	}
	for i := range testCases { //遍历每一个case并执行子测试
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			//创建mock模拟数据库,必须先生成控制器并延迟Finish,控制器是mock的顶层控制
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)
			//构建api请求,创建Server,用httptest创建一个recorder
			server := NewServer(store)
			recorder := httptest.NewRecorder()
			url := fmt.Sprintf("/accounts/%d", tc.accountID)
			request, err := http.NewRequest("GET", url, nil)
			require.NoError(t, err)
			//调用ServeHTTP方法传入recorder和请求,recorder相当于response,body就是bytes.buffer
			server.router.ServeHTTP(recorder, request)
			//接下来需要对比查询到的account和我们生成的account是否一致
			tc.checkResponse(t, recorder)
		})
	}

}

func requireBodyMatchAccount(t *testing.T, body *bytes.Buffer, account db.Account) {
	data, err := ioutil.ReadAll(body)
	require.NoError(t, err)

	var gotAccount db.Account
	err = json.Unmarshal(data, &gotAccount) //从recorder中读出的数据
	require.NoError(t, err)
	require.Equal(t, gotAccount, account) //得到的和输入的值一致

}

func TestCreateAccount(t *testing.T) {
	account := randomAccount()
	testCases := []struct {
		Name          string
		Body          gin.H //便于POST给API
		BuildMock     func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			Name: "OK",
			Body: gin.H{
				"owner":    account.Owner,
				"currency": account.Currency,
			},
			BuildMock: func(store *mockdb.MockStore) {
				arg := db.CreateAccountParams{
					Owner:    account.Owner,
					Balance:  0,
					Currency: account.Currency,
				}
				store.EXPECT().CreateAccount(gomock.Any(), gomock.Eq(arg)).Times(1).
					Return(account, nil)

			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},

		{
			Name: "Internal",
			Body: gin.H{
				"owner":    account.Owner,
				"currency": account.Currency,
			},
			BuildMock: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Account{}, sql.ErrConnDone) //返回一个空的结构体和错误
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			Name: "BadRequest",
			Body: gin.H{
				"currency": util.RandomString(3),
			},
			BuildMock: func(store *mockdb.MockStore) {
				store.EXPECT().CreateAccount(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			store := mockdb.NewMockStore(ctrl)
			tc.BuildMock(store)

			//构建调用api
			server := NewServer(store)
			recorder := httptest.NewRecorder()
			url := "/accounts"
			body, err := json.Marshal(tc.Body)
			require.NoError(t, err)
			request, err := http.NewRequest("POST", url, bytes.NewReader(body))
			require.NoError(t, err)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})

	}

}

func TestListAccount(t *testing.T) {
	//创建5个随机账户
	n := 5
	accounts := make([]db.Account, n)
	for i := 0; i < n; i++ {
		accounts[i] = randomAccount()
	}
	//查询参数
	type Query struct {
		pageID   int
		pageSize int
	}

	testCases := []struct {
		name  string
		query Query
		build func(store *mockdb.MockStore)
		check func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			query: Query{
				pageID:   1,
				pageSize: n,
			},
			build: func(store *mockdb.MockStore) {
				arg := db.ListAccountsParams{
					Limit:  int32(n),
					Offset: 0,
				}
				store.EXPECT().ListAccounts(gomock.Any(), gomock.Eq(arg)).Times(1).
					Return(accounts, nil)
			},
			check: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				//还要检查返回的
				requireBodyMatchAccounts(t, recorder.Body, accounts)
			},
		},
		{
			name: "BadRequest",
			query: Query{
				pageID:   -1,
				pageSize: 1000,
			},
			build: func(store *mockdb.MockStore) {
				store.EXPECT().ListAccounts(gomock.Any(), gomock.Any()).Times(0)
			},
			check: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, recorder.Code, http.StatusBadRequest)
			},
		},
		{
			name: "NotFound",
			query: Query{
				pageID:   1,
				pageSize: n,
			},
			build: func(store *mockdb.MockStore) {
				arg := db.ListAccountsParams{
					Limit:  int32(n),
					Offset: 0,
				}
				store.EXPECT().ListAccounts(gomock.Any(), arg).Times(1).
					Return([]db.Account{}, sql.ErrNoRows)

			},
			check: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, recorder.Code, http.StatusNotFound)
			},
		},
		{
			name: "InternalErr",
			query: Query{
				pageID:   1,
				pageSize: n,
			},
			build: func(store *mockdb.MockStore) {
				arg := db.ListAccountsParams{
					Limit:  int32(n),
					Offset: 0,
				}
				store.EXPECT().ListAccounts(gomock.Any(), arg).Times(1).
					Return([]db.Account{}, sql.ErrConnDone)
			},
			check: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, recorder.Code, http.StatusInternalServerError)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			//构建mock
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			store := mockdb.NewMockStore(ctrl)
			tc.build(store)

			//构建API调用
			server := NewServer(store)
			recorder := httptest.NewRecorder()
			url := fmt.Sprintf("/accounts")
			request, err := http.NewRequest("GET", url, nil)
			require.NoError(t, err)
			//必须将其复制给一个新变量
			q := request.URL.Query()
			q.Add("page_id", strconv.Itoa(tc.query.pageID))
			q.Add("page_size", strconv.Itoa(tc.query.pageSize))
			request.URL.RawQuery = q.Encode()
			server.router.ServeHTTP(recorder, request)

			//检查返回值
			tc.check(recorder)

		})

	}

}

func requireBodyMatchAccounts(t *testing.T, body *bytes.Buffer, accounts []db.Account) {
	data, err := ioutil.ReadAll(body)
	require.NoError(t, err)
	var gotAccounts []db.Account
	err = json.Unmarshal(data, &gotAccounts)
	require.NoError(t, err)
	require.Equal(t, accounts, gotAccounts)
}
