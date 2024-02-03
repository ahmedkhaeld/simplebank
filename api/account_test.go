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
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestGetAccountAPI(t *testing.T) {

	account := randAccount()

	testCases := []struct {
		name          string
		accountID     int64
		buildStubs    func(store *mockdb.MockStore)                           // build test stub that suite the test case
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder) // check response of the api
	}{
		{
			name:      "OK",
			accountID: account.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account.ID)).Times(1).Return(account, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireEqualAccount(t, account, recorder.Body)
			},
		},
		{
			name:      "NotFound",
			accountID: account.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account.ID)).Times(1).Return(db.Account{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:      "InternalServerError",
			accountID: account.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account.ID)).Times(1).Return(db.Account{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:      "InvalidID",
			accountID: 0,
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
			server := NewServer(store)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/api/accounts/%d", tc.accountID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			//Invoke the API Endpoint, and check the response [status code, body]
			server.router.ServeHTTP(recorder, request)

			// check the response
			tc.checkResponse(t, recorder)

		})
	}

}

func randAccount() db.Account {
	return db.Account{
		ID:       util.RandomInt(1, 1000),
		Owner:    util.RandomAccountOwner(),
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
