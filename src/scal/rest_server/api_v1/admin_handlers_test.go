package api_v1

import (
	"testing"

	"scal/settings"
	. "scal/user_lib"

	"github.com/golang/mock/gomock"
	"github.com/ilius/is"
	. "github.com/ilius/ripo"
)

var origAdminEmails = settings.ADMIN_EMAILS

func resetAdminEmails() {
	settings.ADMIN_EMAILS = origAdminEmails
}

func TestAdminGetStats_Unauthenticated(t *testing.T) {
	is := is.New(t)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockReq := NewMockExtendedRequest(ctrl)
	var req ExtendedRequest = mockReq

	mockReq.EXPECT().Header(gomock.Eq("Authorization"))

	res, err := AdminGetStats(req)
	rpcErr := err.(RPCError)
	is.Err(rpcErr)
	is.Equal(rpcErr.Code(), Unauthenticated)
	is.Nil(res)
}

func TestAdminGetStats_NotAdmin(t *testing.T) {
	defer testUserMapClear()

	authorization := "abcd"
	email := "abcd@dummy.com"
	testUserMap = map[string]*UserModel{
		authorization: &UserModel{
			Email: email,
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockReq := NewMockExtendedRequest(ctrl)
	var req ExtendedRequest = mockReq
	mockReq.EXPECT().Header(gomock.Eq("Authorization")).Return(authorization)

	is := is.New(t)
	res, err := AdminGetStats(req)
	rpcErr := err.(RPCError)
	is.Err(rpcErr)
	is.Equal(rpcErr.Code(), PermissionDenied)
	is.Equal(rpcErr.Message(), "you are not admin")
	is.Nil(res)
}

func TestAdminGetStats_EmailNotConfirmed(t *testing.T) {
	defer testUserMapClear()
	defer resetAdminEmails()

	authorization := "abcd"
	email := "abcd@dummy.com"
	testUserMap = map[string]*UserModel{
		authorization: &UserModel{
			Email: email,
		},
	}
	settings.ADMIN_EMAILS = []string{email}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockReq := NewMockExtendedRequest(ctrl)
	var req ExtendedRequest = mockReq
	mockReq.EXPECT().Header(gomock.Eq("Authorization")).Return(authorization)

	is := is.New(t)
	res, err := AdminGetStats(req)
	rpcErr := err.(RPCError)
	is.Equal(rpcErr.Code(), PermissionDenied)
	is.Equal(rpcErr.Message(), "email is not confirmed")
	is.Err(rpcErr)
	is.Nil(res)
}

func TestAdminGetStats_OK(t *testing.T) {
	defer testUserMapClear()
	defer resetAdminEmails()

	authorization := "abcd"
	email := "abcd@dummy.com"
	testUserMap = map[string]*UserModel{
		authorization: &UserModel{
			Email:          email,
			EmailConfirmed: true,
		},
	}
	settings.ADMIN_EMAILS = []string{email}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockReq := NewMockExtendedRequest(ctrl)
	var req ExtendedRequest = mockReq
	mockReq.EXPECT().Header(gomock.Eq("Authorization")).Return(authorization)

	is := is.New(t)
	res, err := AdminGetStats(req)
	is.NotErr(err)
	is.NotNil(res)
	is.NotNil(res.Data)
	dataMap := res.Data.(map[string]interface{})
	locked_resource_count := dataMap["locked_resource_count"]
	is.Equal(locked_resource_count, map[string]int{
		"event":      0,
		"group":      0,
		"user":       0,
		"user_login": 0,
	})
}
