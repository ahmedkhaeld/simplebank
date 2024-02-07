package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	mockdb "github.com/ahmedkhaeld/simplebank/db/mock"
	db "github.com/ahmedkhaeld/simplebank/db/sqlc"
	"github.com/ahmedkhaeld/simplebank/util"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestGetAccountAPI(t *testing.T) {
	user, _ := randomUser(t)
	account := randAccount(user.Username)

	testCases := []struct {
		name          string
		accountID     int64
		authUsername  string
		buildStubs    func(store *mockdb.MockStore)                           // build test stub that suite the test case
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder) // check response of the api
	}{
		{
			name:         "OK",
			accountID:    account.ID,
			authUsername: user.Username,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account.ID)).Times(1).Return(account, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireEqualAccount(t, account, recorder.Body)
			},
		},
		{
			name:         "UnauthorizedUser",
			accountID:    account.ID,
			authUsername: "unauthorized_user",
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(account, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name:         "NoAuthorization",
			accountID:    account.ID,
			authUsername: "",
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name:         "NotFound",
			accountID:    account.ID,
			authUsername: user.Username,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account.ID)).Times(1).Return(db.Account{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:         "InternalServerError",
			accountID:    account.ID,
			authUsername: user.Username,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account.ID)).Times(1).Return(db.Account{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:         "InvalidID",
			accountID:    0,
			authUsername: user.Username,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {

			//Create a Mock Store
			//It creates a controller for managing mocks
			//and initializes a mock database store using NewMockStore.
			crtl := gomock.NewController(t)
			defer crtl.Finish()
			store := mockdb.NewMockStore(crtl)

			//build the stubs
			tc.buildStubs(store)

			//Create HTTP Server and Recorder:
			//It creates an instance of the server
			//and an HTTP response recorder for capturing the server's response.
			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/api/accounts/%d", tc.accountID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			createAndSetAuthToken(t, request, server.tokenMaker, tc.authUsername)

			//Invoke the API Endpoint, and check the response [status code, body]
			server.router.ServeHTTP(recorder, request)

			// check the response
			tc.checkResponse(t, recorder)

		})
	}

}

func TestCreateAccountAPI(t *testing.T) {
	user, _ := randomUser(t)
	account := randAccount(user.Username)

	testCases := []struct {
		name          string
		body          gin.H
		authUsername  string
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"currency": account.Currency,
			},
			authUsername: user.Username,
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.CreateAccountParams{
					Owner:    account.Owner,
					Currency: account.Currency,
					Balance:  0,
				}

				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(account, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusCreated, recorder.Code)
				requireEqualAccount(t, account, recorder.Body)
			},
		},
		{
			name: "NoAuthorization",
			body: gin.H{
				"currency": account.Currency,
			},
			authUsername: "",
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "InternalError",
			body: gin.H{
				"currency": account.Currency,
			},
			authUsername: user.Username,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Account{}, sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "InvalidCurrency",
			body: gin.H{
				"currency": "invalid",
			},
			authUsername: user.Username,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := "/api/accounts"
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(data))
			require.NoError(t, err)

			createAndSetAuthToken(t, request, server.tokenMaker, tc.authUsername)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestListAccountAPI(t *testing.T) {
	user, _ := randomUser(t)

	n := 5
	accounts := make([]db.Account, n)
	for i := 0; i < n; i++ {
		accounts[i] = randAccount(user.Username)
	}

	type Query struct {
		Page  int
		Limit int
	}

	testCases := []struct {
		name          string
		authUsername  string
		query         Query
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			query: Query{
				Page:  1,
				Limit: n,
			},
			authUsername: user.Username,
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.ListAccountsParams{
					Owner:  user.Username,
					Limit:  int32(n),
					Offset: 0,
				}

				store.EXPECT().
					ListAccounts(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(accounts, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireEqualAccounts(t, accounts, recorder.Body)
			},
		},
		{
			name: "InternalError",
			query: Query{
				Page:  1,
				Limit: n,
			},
			authUsername: user.Username,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListAccounts(gomock.Any(), gomock.Any()).
					Times(1).
					Return([]db.Account{}, sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "InvalidPageID",
			query: Query{
				Page:  -1,
				Limit: n,
			},
			authUsername: user.Username,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListAccounts(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "InvalidPageSize",
			query: Query{
				Page:  1,
				Limit: 100000,
			},
			authUsername: user.Username,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListAccounts(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			url := "/api/accounts"
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			// Add query parameters to request URL
			q := request.URL.Query()
			q.Add("page", fmt.Sprintf("%d", tc.query.Page))
			q.Add("limit", fmt.Sprintf("%d", tc.query.Limit))
			request.URL.RawQuery = q.Encode()

			createAndSetAuthToken(t, request, server.tokenMaker, tc.authUsername)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}

}

func randAccount(owner string) db.Account {
	return db.Account{
		ID:       util.RandomInt(1, 1000),
		Owner:    owner,
		Balance:  util.RandomMoney(),
		Currency: util.RandomAccountCurrency(),
	}
}

// requireAccountBodyMatch check the request body matches the account
func requireEqualAccount(t *testing.T, expected db.Account, body *bytes.Buffer) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotData map[string]interface{}
	err = json.Unmarshal(data, &gotData)
	require.NoError(t, err)

	accountData, ok := gotData["Data"].(map[string]interface{})["account"].(map[string]interface{})
	require.True(t, ok, "failed to access the 'account' element in the JSON")
	// Now you can access specific fields within the 'account' element
	id := int(accountData["id"].(float64))
	owner := accountData["owner"].(string)
	balance := float64(accountData["balance"].(float64))
	currency := accountData["currency"].(string)

	// Create an Account structure with the extracted data
	extractedAccount := db.Account{
		ID:       int64(id),
		Owner:    owner,
		Balance:  int64(balance),
		Currency: currency,
	}

	require.Equal(t, expected, extractedAccount)
}
func requireEqualAccounts(t *testing.T, expected []db.Account, body *bytes.Buffer) {
	var actual []db.Account
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotData map[string]interface{}
	err = json.Unmarshal(data, &gotData)
	require.NoError(t, err)

	gotAccounts, ok := gotData["Data"].(map[string]interface{})["accounts"]
	require.True(t, ok, "failed to access the 'accounts' element in the JSON")

	// Iterate through the list of accounts and create db.Account objects
	for _, accountData := range gotAccounts.([]interface{}) {
		accountMap := accountData.(map[string]interface{})

		id := int(accountMap["id"].(float64))
		owner := accountMap["owner"].(string)
		balance := float64(accountMap["balance"].(float64))
		currency := accountMap["currency"].(string)

		actual = append(actual, db.Account{
			ID:       int64(id),
			Owner:    owner,
			Balance:  int64(balance),
			Currency: currency,
		})
	}

	// Assert that the actual accounts match the expected ones
	require.Equal(t, expected, actual)
}
