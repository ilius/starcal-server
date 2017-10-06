package settings

import (
	"fmt"

	jwt "github.com/dgrijalva/jwt-go"
)

func init() {
	jwtSigningMethod := jwt.GetSigningMethod(JWT_ALG)
	if jwtSigningMethod == nil {
		panic(fmt.Sprintf("invalid settings.JWT_ALG = %#v", JWT_ALG))
	}
	if JWT_TOKEN_SECRET == "" {
		panic("settings.JWT_TOKEN_SECRET can not be empty, build again")
	}
	if CONFIRM_EMAIL_SECRET == "" {
		panic("settings.CONFIRM_EMAIL_SECRET can not be empty, build again")
	}
}
