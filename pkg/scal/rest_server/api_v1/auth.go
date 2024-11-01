package api_v1

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/ilius/starcal-server/pkg/scal/settings"
	"github.com/ilius/starcal-server/pkg/scal/storage"
	"github.com/ilius/starcal-server/pkg/scal/user_lib"

	"github.com/alexandrevicenzi/unchained"
	jwt "github.com/golang-jwt/jwt/v5"
	rp "github.com/ilius/ripo"
)

const TOKEN_CONTEXT = "user"

var (
	errTokenNotFound  = errors.New("JWT Authorization token not found")
	errTokenBadFormat = errors.New("JWT Authorization header format must be 'Bearer {token}'")
	errTokenInvalid   = errors.New("JWT token is invalid")
	errClaimsNotFound = errors.New("JWT claims not found")

	// errClaimsEmailNotFound = errors.New("Email not found in JWT claims")
)

// testUserMap: a map to set in test, to bypass JWY authentication
// key of map is the value of req.Header("Authorization")
var testUserMap map[string]*user_lib.UserModel

func testUserMapClear() {
	testUserMap = nil
}

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
		func(token *jwt.Token) (any, error) {
			return []byte(settings.JWT_TOKEN_SECRET), nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("error parsing token: %v", err)
	}

	expectedAlg := settings.JWT_ALG
	tokenAlg := token.Header["alg"]
	if expectedAlg != tokenAlg {
		return nil, fmt.Errorf(
			"expected %s signing method but token specified %s",
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

func AuthError(err error) rp.RPCError {
	randomSleep(4)
	return rp.NewError(rp.Unauthenticated, "", err)
}

func ForbiddenError(publicMsg string, err error) rp.RPCError {
	randomSleep(4)
	return rp.NewError(rp.PermissionDenied, publicMsg, err)
}

func isGoTest() bool {
	return strings.HasSuffix(os.Args[0], ".test")
}

func CheckAuth(req rp.Request) (*user_lib.UserModel, error) {
	log.Debug(req.HandlerName())
	authHeader := req.Header("Authorization")
	if authHeader == "" {
		return nil, AuthError(errTokenNotFound)
	}
	if isGoTest() && testUserMap != nil {
		userModel := testUserMap[authHeader]
		if userModel != nil {
			return userModel, nil
		}
	}
	token, err := TokenFromHeader(authHeader)
	if err != nil {
		return nil, AuthError(err)
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, rp.NewError(rp.Internal, "", errClaimsNotFound)
	}
	emailI, ok := claims["email"]
	if !ok {
		return nil, AuthError(fmt.Errorf("error parsing token: Missing 'email'"))
	}
	email, ok := emailI.(string)
	if !ok {
		return nil, AuthError(fmt.Errorf("error parsing token: Bad 'email'"))
	}
	if email == "" {
		return nil, AuthError(fmt.Errorf("error parsing token: Empty 'email'"))
	}
	db, err := storage.GetDB()
	if err != nil {
		return nil, rp.NewError(rp.Internal, "", err)
	}
	userModel := user_lib.UserModelByEmail(email, db)
	if userModel == nil {
		// SetHttpErrorUserNotFound(w, email) // FIXME
		return nil, AuthError(fmt.Errorf("error parsing token: Bad 'email'"))
	}
	if userModel.Locked {
		return nil, ForbiddenError("user is locked", nil)
	}
	issuedAtI, ok := claims["iat"]
	if !ok {
		return nil, AuthError(fmt.Errorf("error parsing token: Missing 'iat'"))
	}
	issuedAtFloat, ok := issuedAtI.(float64)
	if !ok {
		return nil, AuthError(fmt.Errorf("error parsing token: Bad 'iat'"))
	}
	issuedAt := time.Unix(int64(issuedAtFloat), 0)
	if userModel.LastLogoutTime != nil {
		if userModel.LastLogoutTime.After(issuedAt) {
			return nil, AuthError(fmt.Errorf("error parsing token: Token is expired"))
		}
	}
	userModel.TokenIssuedAt = &issuedAt
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

func EmailIsAdmin(email string) bool {
	for _, adminEmail := range settings.ADMIN_EMAILS {
		if email == adminEmail {
			return true
		}
	}
	return false
}

func AdminCheckAuth(req rp.Request) (*user_lib.UserModel, error) {
	userModel, err := CheckAuth(req)
	if err != nil {
		return nil, err
	}
	if !EmailIsAdmin(userModel.Email) {
		return nil, rp.NewError(rp.PermissionDenied, "you are not admin", nil)
	}
	if !userModel.EmailConfirmed {
		return nil, rp.NewError(rp.PermissionDenied, "email is not confirmed", nil)
	}
	return userModel, nil
}

// first arg is email, will we need it or should I remove it?
func GetPasswordHash(_ string, password string) (string, error) {
	return unchained.MakePassword(
		password,
		"", // BCrypt does not take salt as input
		unchained.BCryptHasher,
	)
}

// first arg is email, will we need it or should I remove it?
func CheckPasswordHash(_ string, password string, pwHash string) bool {
	valid, err := unchained.CheckPassword(
		password,
		pwHash,
	)
	if err != nil {
		log.Error("error in CheckPassword: %v", err)
	}
	return valid
}

func NewSignedToken(userModel *user_lib.UserModel) (string, time.Time) {
	now := time.Now()
	exp := now.Add(settings.JWT_TOKEN_EXP_SECONDS * time.Second).UTC()
	token := jwt.NewWithClaims(
		jwt.GetSigningMethod(settings.JWT_ALG),
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
	return signedToken, exp
}
