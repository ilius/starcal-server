package api_v1

import (
	"testing"

	"github.com/ilius/starcal-server/pkg/scal/settings"
	. "github.com/ilius/starcal-server/pkg/scal/user_lib"

	"github.com/golang/mock/gomock"
	"github.com/ilius/is"
	. "github.com/ilius/ripo"
)

var origAdminEmails = settings.ADMIN_EMAILS

func resetAdminEmails() {
	settings.ADMIN_EMAILS = origAdminEmails
}

func testAdmin_Unauthenticated(t *testing.T, handler Handler) {
	is := is.New(t)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockReq := NewMockExtendedRequest(ctrl)
	var req ExtendedRequest = mockReq

	mockReq.EXPECT().Header(gomock.Eq("Authorization"))

	res, err := handler(req)
	rpcErr := err.(RPCError)
	is.Err(rpcErr)
	is.Equal(rpcErr.Code(), Unauthenticated)
	is.Nil(res)
}

func testAdmin_NotAdmin(t *testing.T, handler Handler) {
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
	res, err := handler(req)
	rpcErr := err.(RPCError)
	is.Err(rpcErr)
	is.Equal(rpcErr.Code(), PermissionDenied)
	is.Equal(rpcErr.Message(), "you are not admin")
	is.Nil(res)
}

func testAdmin_EmailNotConfirmed(t *testing.T, handler Handler) {
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
	res, err := handler(req)
	rpcErr := err.(RPCError)
	is.Equal(rpcErr.Code(), PermissionDenied)
	is.Equal(rpcErr.Message(), "email is not confirmed")
	is.Err(rpcErr)
	is.Nil(res)
}

func testAdmin_OK(t *testing.T, handler Handler) *Response {
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
	res, err := handler(req)
	is.NotErr(err)
	is.NotNil(res)

	return res
}

func TestAdminGetStats(t *testing.T) {
	handler := AdminGetStats

	testAdmin_Unauthenticated(t, handler)
	testAdmin_NotAdmin(t, handler)
	testAdmin_EmailNotConfirmed(t, handler)

	res := testAdmin_OK(t, handler)

	is := is.New(t)
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

func TestAdminListLockedResources(t *testing.T) {
	handler := AdminListLockedResources

	testAdmin_Unauthenticated(t, handler)
	testAdmin_NotAdmin(t, handler)
	testAdmin_EmailNotConfirmed(t, handler)

	res := testAdmin_OK(t, handler)

	is := is.New(t)
	is.NotNil(res.Data)
	is.Equal(res.Data, map[string][]string{
		"event":      {},
		"group":      {},
		"user":       {},
		"user_login": {},
	})

}
