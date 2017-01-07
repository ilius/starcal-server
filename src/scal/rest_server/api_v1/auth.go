package api_v1

import (
	"crypto/sha512"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"net/http"
	"scal/settings"
	. "scal/user_lib"
	"strings"
	"time"
)

const TOKEN_CONTEXT = "user"

var JWT_SIGNING_METHOD = jwt.SigningMethodHS256

var (
	ErrTokenNotFound       = errors.New("JWT Authorization token not found")
	ErrTokenBadFormat      = errors.New("JWT Authorization header format must be 'Bearer {token}'")
	ErrTokenInvalid        = errors.New("JWT token is invalid")
	ErrClaimsNotFound      = errors.New("JWT claims not found")
	ErrClaimsEmailNotFound = errors.New("Email not found in JWT claims")
)

func init() {
	if settings.JWT_TOKEN_SECRET == "" {
		panic("settings.JWT_TOKEN_SECRET can not be empty, build again")
	}
}

func ExtractToken(r *http.Request) (*jwt.Token, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, ErrTokenNotFound
	}
	authHeaderParts := strings.Split(authHeader, " ")
	if len(authHeaderParts) != 2 || strings.ToLower(authHeaderParts[0]) != "bearer" {
		return nil, ErrTokenBadFormat
	}
	tokenStr := authHeaderParts[1]
	if tokenStr == "" {
		return nil, ErrTokenNotFound
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

	expectedAlg := JWT_SIGNING_METHOD.Alg()
	tokenAlg := token.Header["alg"]
	if expectedAlg != tokenAlg {
		return nil, fmt.Errorf(
			"Expected %s signing method but token specified %s",
			expectedAlg,
			tokenAlg,
		)
	}

	if !token.Valid {
		return nil, ErrTokenInvalid
	}

	return token, nil
}

func CheckAuthGetUserModel(w http.ResponseWriter, r *http.Request) *UserModel {
	token, err := ExtractToken(r)
	if err != nil {
		SetHttpError(w, http.StatusUnauthorized, err.Error())
		return nil
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		SetHttpErrorInternal(w, ErrClaimsNotFound)
		return nil
	}
	emailI, ok := claims["email"]
	if !ok {
		SetHttpError(
			w,
			http.StatusUnauthorized,
			"Error parsing token: Missing 'email'",
		)
		return nil
	}
	email, ok := emailI.(string)
	if !ok {
		SetHttpError(
			w,
			http.StatusUnauthorized,
			"Error parsing token: Bad 'email'",
		)
		return nil
	}
	if email == "" {
		SetHttpError(
			w,
			http.StatusUnauthorized,
			"Error parsing token: Empty 'email'",
		)
		return nil
	}
	userModel := UserModelByEmail(email, globalDb)
	if userModel == nil {
		SetHttpError(
			w,
			http.StatusUnauthorized,
			"Error parsing token: Bad 'email'",
		)
		//SetHttpErrorUserNotFound(w, email) // FIXME
		return nil
	}
	if userModel.Locked {
		SetHttpError(
			w,
			http.StatusForbidden,
			"Error parsing token: User is locked",
		)
		return nil
	}
	if userModel.LastLogoutTime != nil {
		issuedAtI, ok := claims["iat"]
		if !ok {
			SetHttpError(
				w,
				http.StatusUnauthorized,
				"Error parsing token: Missing 'iat'",
			)
			return nil
		}
		issuedAt, err := time.Parse(time.RFC3339, issuedAtI.(string))
		if err != nil {
			SetHttpError(
				w,
				http.StatusUnauthorized,
				"Error parsing token: Bad 'iat'",
			)
			return nil
		}
		if userModel.LastLogoutTime.After(issuedAt) {
			SetHttpError(
				w,
				http.StatusUnauthorized,
				"Error parsing token: Token is expired",
			)
			return nil
		}
	}
	return userModel
}

/*
NEW:
	userModel := CheckAuthGetUserModel(w, r)
	if userModel == nil {
		return
	}
	email := userModel.Email

OLD:
	ok, email := CheckAuthGetEmail(w, r)
	if !ok {
		return
	}
*/

func authWrap(protectedPage http.HandlerFunc) http.HandlerFunc { // REMOVE, FIXME
	return protectedPage
}

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
	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		jwt.MapClaims{
			"email": userModel.Email, // FIXME
			"iat":   now,
			"exp": now.Add(
				settings.JWT_TOKEN_EXP_SECONDS * time.Second,
			).Unix(),
			/*jwt.StandardClaims {
				//ExpiresAt: expireToken.Unix(),
				Issuer:	settings.HOST, // add ":port" too? FIXME
			},*/
		},
	)

	// Signs the token with a secret.
	signedToken, _ := token.SignedString([]byte(
		settings.JWT_TOKEN_SECRET,
	))
	return signedToken
}
