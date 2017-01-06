package api_v1

import (
	"crypto/sha512"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"gopkg.in/mgo.v2/bson"
	"io/ioutil"
	"net"
	"net/http"
	"scal"
	"scal/event_lib"
	"scal/settings"
	"scal/storage"
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

func CheckAuthGetEmail(w http.ResponseWriter, r *http.Request) (bool, string) {
	token, err := ExtractToken(r)
	if err != nil {
		SetHttpError(w, http.StatusUnauthorized, err.Error())
		return false, ""
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		SetHttpErrorInternal(w, ErrClaimsNotFound)
		return false, ""
	}
	emailI, ok := claims["email"]
	if !ok {
		SetHttpError(w, http.StatusUnauthorized, "missing email")
		return false, ""
	}
	email, ok := emailI.(string)
	if !ok {
		SetHttpError(w, http.StatusUnauthorized, "bad email")
		return false, ""
	}
	if email == "" {
		SetHttpError(w, http.StatusUnauthorized, "empty email")
		return false, ""
	}
	return true, email
}

/*
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

func Login(w http.ResponseWriter, r *http.Request) {
	// Expires the token and cookie in 30 days
	//expireToken := time.Now().Add(30 * time.Day)
	//expireCookie := time.Now().Add(30 * time.Day)

	authModel := struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}{}
	body, _ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(body, &authModel)
	if err != nil {
		msg := err.Error()
		//if strings.Contains(msg, "") {
		//	msg = ""
		//}
		SetHttpError(w, http.StatusBadRequest, msg)
		return
	}
	if authModel.Email == "" {
		SetHttpError(w, http.StatusBadRequest, "missing 'email'")
		return
	}
	if authModel.Password == "" {
		SetHttpError(w, http.StatusBadRequest, "missing 'password'")
		return
	}

	db, err := storage.GetDB()
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}

	userModel := UserModelByEmail(authModel.Email, db)
	if userModel == nil {
		SetHttpError(
			w,
			http.StatusForbidden,
			"authentication failed",
		)
		return
	}

	if GetPasswordHash(authModel.Email, authModel.Password) != userModel.Password {
		SetHttpError(
			w,
			http.StatusForbidden,
			"authentication failed",
		)
		return
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		jwt.MapClaims{
			"email": userModel.Email, // FIXME
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

	/*
		// Place the token in the client's cookie
		cookie := http.Cookie{
			Name:  "Auth",
			Value: signedToken,
			//Expires: expireCookie,
			HttpOnly: true,
		}
		http.SetCookie(w, &cookie)
	*/

	json.NewEncoder(w).Encode(scal.M{
		"token": signedToken,
	})
}

func RegisterUser(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	userModel := UserModel{}
	remoteIp, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}

	body, _ := ioutil.ReadAll(r.Body)
	err = json.Unmarshal(body, &userModel)
	if err != nil {
		msg := err.Error()
		//if strings.Contains(msg, "") {
		//	msg = ""
		//}
		SetHttpError(w, http.StatusBadRequest, msg)
		return
	}
	if userModel.Email == "" {
		SetHttpError(w, http.StatusBadRequest, "missing 'email'")
		return
	}
	if userModel.Password == "" {
		SetHttpError(w, http.StatusBadRequest, "missing 'password'")
		return
	}
	db, err := storage.GetDB()
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}
	anotherUserModel := UserModelByEmail(userModel.Email, db)
	if anotherUserModel != nil {
		SetHttpError(w, http.StatusBadRequest, "duplicate 'email'")
		return
	}

	// add new field userModel.PasswordHash, FIXME
	userModel.Password = GetPasswordHash(
		userModel.Email,
		userModel.Password,
	)
	defaultGroup := event_lib.EventGroupModel{
		Id:         bson.NewObjectId(),
		Title:      userModel.Email,
		OwnerEmail: userModel.Email,
	}
	err = db.Insert(defaultGroup)
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}
	userModel.DefaultGroupId = &defaultGroup.Id
	err = db.Insert(UserChangeLogModel{
		Time:         time.Now(),
		RequestEmail: "", // FIXME
		RemoteIp:     remoteIp,
		FuncName:     "RegisterUser",
		Email: &[2]*string{
			nil,
			&userModel.Email,
		},
		DefaultGroupId: &[2]*bson.ObjectId{
			nil,
			userModel.DefaultGroupId,
		},
		//FullName: &[2]*string{
		//	nil
		//	&userModel.FullName,
		//},
	})
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}

	err = db.Insert(userModel)
	if err != nil {
		SetHttpErrorInternal(w, err)
		return
	}
	json.NewEncoder(w).Encode(scal.M{})
}

func init() {
	if settings.JWT_TOKEN_SECRET == "" {
		panic("settings.JWT_TOKEN_SECRET can not be empty, build again")
	}
	routeGroups = append(routeGroups, RouteGroup{
		Base: "auth",
		Map: RouteMap{
			"RegisterUser": {
				"POST",
				"register",
				RegisterUser,
			},
			"Login": {
				"POST",
				"login",
				Login,
			},
		},
	})
}
