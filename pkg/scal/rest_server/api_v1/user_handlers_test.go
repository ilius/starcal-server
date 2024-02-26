package api_v1

import (
	"testing"
	"time"

	"github.com/ilius/starcal-server/pkg/scal/user_lib"

	"github.com/ilius/ripo"
)

func TestUserFullName(t *testing.T) {
	email := "a1@dummy.com"
	fullName := "John Doe"

	h := NewTestHelper(t, email)
	defer h.Finish()
	h.Start()

	is := h.Is()
	is.Equal(h.UserModel().FullName, "")

	{
		req, mockReq := h.NewRequestMock(true, "127.0.0.1")
		mockReq.EXPECT().GetString("fullName").Return(&fullName, nil)
		res, err := SetUserFullName(req)
		is.NotErr(err)
		is.NotNil(res)
		is.Equal(h.UserModel().FullName, fullName)
	}
	{
		req, _ := h.NewRequestMock(true, "127.0.0.1")
		res, err := UnsetUserFullName(req)
		is.NotErr(err)
		is.NotNil(res)
		is.Equal(h.UserModel().FullName, "")
	}
}

func TestUserDefaultGroupId(t *testing.T) {
	email := "a2@dummy.com"
	groupTitle := "group 1"

	h := NewTestHelper(t, email)
	defer h.Finish()
	h.Start()

	is := h.Is()
	is.Equal(*h.UserModel().DefaultGroupId, h.DefaultGroup().Id)

	groupModel := h.CreateGroup(groupTitle)
	is.NotEqual(groupModel.Id, h.DefaultGroup().Id)

	{
		req, mockReq := h.NewRequestMock(true, "127.0.0.1")
		mockReq.EXPECT().GetString("defaultGroupId").Return(&groupModel.Id, nil)
		res, err := SetUserDefaultGroupId(req)
		is.NotErr(err)
		is.NotNil(res)
		is.Equal(*h.UserModel().DefaultGroupId, groupModel.Id)
	}
	{
		req, _ := h.NewRequestMock(true, "127.0.0.1")
		res, err := UnsetUserDefaultGroupId(req)
		is.NotErr(err)
		is.NotNil(res)
		is.Nil(h.UserModel().DefaultGroupId)
	}
}

func TestGetUserInfo(t *testing.T) {
	email := "a3@dummy.com"

	h := NewTestHelper(t, email)
	defer h.Finish()
	h.Start()

	is := h.Is()

	{
		req, _ := h.NewRequestMock(true, "")
		res, err := GetUserInfo(req)
		if err != nil {
			rpcErr := err.(ripo.RPCError)
			t.Fatal(rpcErr.Code(), ":", rpcErr.Cause())
		}
		is.NotNil(res)
		dataMap := res.Data.(map[string]any)
		is.Equal(dataMap["email"], email)
		defaultGroupId := dataMap["defaultGroupId"].(*string)
		is.NotNil(defaultGroupId)
		is.Equal(*defaultGroupId, h.DefaultGroup().Id)
		is.Equal(dataMap["fullName"], "")
		is.Equal(dataMap["emailConfirmed"], false)
		is.Nil(dataMap["lastLogoutTime"])
		lastLogins := dataMap["lastLogins"].([]*user_lib.UserLoginAttemptModel)
		is.Equal(len(lastLogins), 0)
	}
}

func addLoginHistory(h *TestHelper, count int, sleep time.Duration, remoteIp string) {
	db := h.DB()
	userModel := h.UserModel()
	for range count {
		err := db.Insert(&user_lib.UserLoginAttemptModel{
			Time:       time.Now(),
			UserId:     userModel.Id,
			Email:      userModel.Email,
			RemoteIp:   remoteIp,
			Successful: true,
		})
		if err != nil {
			panic(err)
		}
		time.Sleep(sleep)
	}
}

func TestGetUserLoginHistoryEmpty(t *testing.T) {
	email := "a4@dummy.com"

	h := NewTestHelper(t, email)
	defer h.Finish()
	h.Start()

	is := h.Is()
	{
		req, mockReq := h.NewRequestMock(true, "")
		mockReq.EXPECT().GetIntDefault("limit", 20).Return(20, nil)
		res, err := GetUserLoginHistory(req)
		if err != nil {
			rpcErr := err.(ripo.RPCError)
			t.Fatal(rpcErr.Code(), ":", rpcErr.Cause())
		}
		is.NotNil(res)
		dataMap := res.Data.(map[string]any)
		lastLogins := dataMap["lastLogins"].([]*user_lib.UserLoginAttemptModel)
		is.Equal(len(lastLogins), 0)
	}
}

func TestGetUserLoginHistoryFull(t *testing.T) {
	email := "a5@dummy.com"
	remoteIp := "127.0.0.1"
	loginsCount := 15
	sleep := 100 * time.Millisecond

	h := NewTestHelper(t, email)
	defer h.Finish()
	defer h.RemoveLoginHistory()
	h.Start()

	addLoginHistory(h, loginsCount, sleep, remoteIp)

	is := h.Is()
	userId := h.UserModel().Id
	{
		req, mockReq := h.NewRequestMock(true, "")
		mockReq.EXPECT().GetIntDefault("limit", 20).Return(20, nil)
		res, err := GetUserLoginHistory(req)
		if err != nil {
			rpcErr := err.(ripo.RPCError)
			t.Fatal(rpcErr.Code(), ":", rpcErr.Cause())
		}
		is.NotNil(res)
		dataMap := res.Data.(map[string]any)
		lastLogins := dataMap["lastLogins"].([]*user_lib.UserLoginAttemptModel)
		is.Equal(len(lastLogins), loginsCount)
		for _, m := range lastLogins {
			is.Equal(m.UserId, userId)
			is.Equal(m.Email, email)
			is.Equal(m.RemoteIp, remoteIp)
			is.True(m.Successful)
		}
	}
}

func TestGetUserLoginHistoryLimit1(t *testing.T) {
	email := "a6@dummy.com"
	remoteIp := "127.0.0.1"
	loginsCount := 25
	sleep := 100 * time.Millisecond

	h := NewTestHelper(t, email)
	defer h.Finish()
	defer h.RemoveLoginHistory()
	h.Start()

	addLoginHistory(h, loginsCount, sleep, remoteIp)

	is := h.Is()
	userId := h.UserModel().Id
	{
		req, mockReq := h.NewRequestMock(true, "")
		mockReq.EXPECT().GetIntDefault("limit", 20).Return(20, nil)
		res, err := GetUserLoginHistory(req)
		if err != nil {
			rpcErr := err.(ripo.RPCError)
			t.Fatal(rpcErr.Code(), ":", rpcErr.Cause())
		}
		is.NotNil(res)
		dataMap := res.Data.(map[string]any)
		lastLogins := dataMap["lastLogins"].([]*user_lib.UserLoginAttemptModel)
		is.Equal(len(lastLogins), 20)
		for _, m := range lastLogins {
			is.Equal(m.UserId, userId)
			is.Equal(m.Email, email)
			is.Equal(m.RemoteIp, remoteIp)
			is.True(m.Successful)
		}
	}
}

func TestGetUserLoginHistoryLimit2(t *testing.T) {
	email := "a7@dummy.com"
	remoteIp := "127.0.0.1"
	loginsCount := 10
	sleep := 100 * time.Millisecond

	h := NewTestHelper(t, email)
	defer h.Finish()
	defer h.RemoveLoginHistory()
	h.Start()

	addLoginHistory(h, loginsCount, sleep, remoteIp)

	is := h.Is()
	userId := h.UserModel().Id
	{
		req, mockReq := h.NewRequestMock(true, "")
		mockReq.EXPECT().GetIntDefault("limit", 20).Return(7, nil)
		res, err := GetUserLoginHistory(req)
		if err != nil {
			rpcErr := err.(ripo.RPCError)
			t.Fatal(rpcErr.Code(), ":", rpcErr.Cause())
		}
		is.NotNil(res)
		dataMap := res.Data.(map[string]any)
		lastLogins := dataMap["lastLogins"].([]*user_lib.UserLoginAttemptModel)
		is.Equal(len(lastLogins), 7)
		for _, m := range lastLogins {
			is.Equal(m.UserId, userId)
			is.Equal(m.Email, email)
			is.Equal(m.RemoteIp, remoteIp)
			is.True(m.Successful)
		}
	}
}
