package api_v1

import (
	"crypto/sha512"
	"errors"
	"fmt"
	"math/rand"
	"scal/settings"
	. "scal/user_lib"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	. "github.com/ilius/restpc"
)

const TOKEN_CONTEXT = "user"

var (
	errTokenNotFound       = errors.New("JWT Authorization token not found")
	errTokenBadFormat      = errors.New("JWT Authorization header format must be 'Bearer {token}'")
	errTokenInvalid        = errors.New("JWT token is invalid")
	errClaimsNotFound      = errors.New("JWT claims not found")
	errClaimsEmailNotFound = errors.New("Email not found in JWT claims")
)

func TokenFromHeader(authHeader string) (*jwt.Token, error) {
	authHeaderParts := strings.Split(authHeader, " ")
	if len(authHeaderParts) != 2 || strings.ToLower(authHeaderParts[0]) != "bearer" {
		return nil, errTokenBadFormat
	}
	tokenStr := authHeaderParts[1]
	if tokenStr == "" {
		return nil, errTokenNotFound
	}

	token, err := jwt.Parse(
		tokenStr,
		func(token *jwt.Token) (interface{}, error) {
			return []byte(settings.JWT_TOKEN_SECRET), nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("Error parsing token: %v", err)
	}

	expectedAlg := settings.JWT_ALG
	tokenAlg := token.Header["alg"]
	if expectedAlg != tokenAlg {
		return nil, fmt.Errorf(
			"Expected %s signing method but token specified %s",
			expectedAlg,
			tokenAlg,
		)
	}

	if !token.Valid {
		return nil, errTokenInvalid
	}

	return token, nil
}

func randomSleep(maxSeconds int) {
	maxMS := 1000 * maxSeconds
	minMS := maxMS / 2
	time.Sleep(time.Duration(minMS+rand.Intn(maxMS-minMS)) * time.Millisecond)
}

func AuthError(err error) RPCError {
	randomSleep(4)
	return NewError(Unauthenticated, "", err)
}

func ForbiddenError(publicMsg string, err error) RPCError {
	randomSleep(4)
	return NewError(PermissionDenied, publicMsg, err)
}

func CheckAuth(req Request) (*UserModel, error) {
	authHeader := req.GetHeader("Authorization")
	if authHeader == "" {
		return nil, AuthError(errTokenNotFound)
	}
	token, err := TokenFromHeader(authHeader)
	if err != nil {
		return nil, AuthError(err)
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, NewError(Internal, "", errClaimsNotFound)
	}
	emailI, ok := claims["email"]
	if !ok {
		return nil, AuthError(fmt.Errorf("Error parsing token: Missing 'email'"))
	}
	email, ok := emailI.(string)
	if !ok {
		return nil, AuthError(fmt.Errorf("Error parsing token: Bad 'email'"))
	}
	if email == "" {
		return nil, AuthError(fmt.Errorf("Error parsing token: Empty 'email'"))
	}
	userModel := UserModelByEmail(email, globalDb)
	if userModel == nil {
		//SetHttpErrorUserNotFound(w, email) // FIXME
		return nil, AuthError(fmt.Errorf("Error parsing token: Bad 'email'"))
	}
	if userModel.Locked {
		return nil, ForbiddenError("user is locked", nil)
	}
	if userModel.LastLogoutTime != nil {
		issuedAtI, ok := claims["iat"]
		if !ok {
			return nil, AuthError(fmt.Errorf("Error parsing token: Missing 'iat'"))
		}
		issuedAt, err := time.Parse(time.RFC3339, issuedAtI.(string))
		if err != nil {
			return nil, AuthError(fmt.Errorf("Error parsing token: Bad 'iat'"))
		}
		if userModel.LastLogoutTime.After(issuedAt) {
			return nil, AuthError(fmt.Errorf("Error parsing token: Token is expired"))
		}
	}
	return userModel, nil
}

/*
NEW:
	userModel, err := CheckAuth(req)
	if err != nil {
		return nil, err
	}
	email := userModel.Email

OLD:
	userModel := CheckAuthGetUserModel(w, r)
	if userModel == nil {
		return
	}
	email := userModel.Email

VERY OLD:
	ok, email := CheckAuthGetEmail(w, r)
	if !ok {
		return
	}
*/

func GetPasswordHash(email string, password string) string {
	return fmt.Sprintf(
		"%x",
		sha512.Sum512(
			[]byte(
				fmt.Sprintf(
					"%s:%s:%s",
					email,
					settings.PASSWORD_SALT,
					password,
				),
			),
		),
	)
}

func NewSignedToken(userModel *UserModel) string {
	now := time.Now()
	exp := now.Add(settings.JWT_TOKEN_EXP_SECONDS * time.Second)
	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		jwt.MapClaims{
			"email": userModel.Email,
			"iat":   now.Unix(),
			"exp":   exp.Unix(),
		},
		// jwt.StandardClaims {
		// 	//ExpiresAt: exp.Unix(),
		// 	Issuer:	settings.HOST, // add ":port" too? FIXME
		// },
	)

	// Signs the token with a secret.
	signedToken, _ := token.SignedString([]byte(
		settings.JWT_TOKEN_SECRET,
	))
	return signedToken
}
