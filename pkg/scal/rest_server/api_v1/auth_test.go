package api_v1

import (
	"testing"
	"time"

	"github.com/ilius/starcal-server/pkg/scal/settings"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/ilius/is"
	"github.com/ilius/libgostarcal/utils"
	"github.com/ilius/ripo"
)

func TestIsGoTest(t *testing.T) {
	is := is.New(t)
	is.True(isGoTest())
}

func TestRegisterUser(t *testing.T) {
	email := "a8@dummy.com"
	remoteIp := "127.0.0.1"
	password, err := utils.GenerateRandomBase64String(16)
	if err != nil {
		panic(err)
	}

	h := NewTestHelper(t, "") // not passing email so it does not create user
	defer h.Finish()
	defer h.DeleteUserByEmail(email)
	h.Start()

	is := h.Is()
	{
		req, mockReq := h.NewRequestMock(false, remoteIp)
		mockReq.EXPECT().GetString("email", ripo.FromBody).Return(&email, nil)
		mockReq.EXPECT().GetString("password", ripo.FromBody).Return(&password, nil)
		mockReq.EXPECT().Host().Return("127.0.0.1")
		res, err := RegisterUser(req)
		if err != nil {
			rpcErr := err.(ripo.RPCError)
			t.Fatal(rpcErr.Code(), ":", rpcErr.Cause())
		}
		is.NotNil(res)
		dataMap := res.Data.(map[string]interface{})
		t.Log(dataMap)
	}
	{
		req, mockReq := h.NewRequestMock(false, remoteIp)
		mockReq.EXPECT().GetString("email", ripo.FromBody).Return(&email, nil)
		mockReq.EXPECT().GetString("password", ripo.FromBody).Return(&password, nil)
		res, err := RegisterUser(req)
		is.NotNil(err)
		rpcErr := err.(ripo.RPCError)
		is.Equal(rpcErr.Code(), ripo.AlreadyExists)
		is.Equal(rpcErr.Message(), "email is already registered")
		is.Nil(res)
	}
}

func TestLogin(t *testing.T) {
	email := "a9@dummy.com"
	remoteIp := "127.0.0.1"
	password, err := utils.GenerateRandomBase64String(16)
	if err != nil {
		panic(err)
	}

	h := NewTestHelper(t, email)
	defer h.Finish()
	h.SetUserPassword(password)
	h.Start()

	is := h.Is()
	{
		req, mockReq := h.NewRequestMock(false, remoteIp)
		mockReq.EXPECT().GetString("email", ripo.FromBody).Return(&email, nil)
		mockReq.EXPECT().GetString("password", ripo.FromBody).Return(&password, nil)

		now := time.Now()
		res, err := Login(req)
		if err != nil {
			rpcErr := err.(ripo.RPCError)
			t.Fatal(rpcErr.Code(), ":", rpcErr.Cause())
		}
		is.NotNil(res)
		dataMap := res.Data.(map[string]interface{})
		expETA := now.Add(settings.JWT_TOKEN_EXP_SECONDS * time.Second)
		{
			expStr := dataMap["expiration"].(string)
			exp, err := time.Parse(time.RFC3339, expStr)
			is.NotErr(err)
			expDelta := exp.Sub(expETA)
			is.Msg(
				"exp=%v, expETA=%v, expDelta=%v",
				exp, expETA, expDelta,
			).True(expDelta < 2*time.Second)
		}
		{
			tokenStr := dataMap["token"].(string)
			_, err := jwt.Parse(
				tokenStr,
				func(token *jwt.Token) (interface{}, error) {
					return []byte(settings.JWT_TOKEN_SECRET), nil
				},
			)
			is.NotErr(err)
		}
	}
}
