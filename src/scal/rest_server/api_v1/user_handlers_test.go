package api_v1

import (
	"scal/user_lib"
	"testing"

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
		dataMap := res.Data.(map[string]interface{})
		is.Equal(dataMap["email"], email)
		is.Equal(dataMap["defaultGroupId"], h.DefaultGroup().Id)
		is.Equal(dataMap["fullName"], "")
		is.Equal(dataMap["emailConfirmed"], false)
		is.Nil(dataMap["lastLogoutTime"])
		lastLogins := dataMap["lastLogins"].([]*user_lib.UserLoginAttemptModel)
		is.Equal(len(lastLogins), 0)
	}
}
