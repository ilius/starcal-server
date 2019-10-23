package api_v1

import (
	"testing"
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
