package event_lib

import (
	"fmt"
	"scal/settings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"gopkg.in/mgo.v2/bson"
)

func init() {
	jwtSigningMethod := jwt.GetSigningMethod(settings.EVENT_INVITE_TOKEN_ALG)
	if jwtSigningMethod == nil {
		panic(fmt.Sprintf("invalid settings.EVENT_INVITE_TOKEN_ALG = %#v", settings.EVENT_INVITE_TOKEN_ALG))
	}
}

// CheckEventInvitationToken: returns (&email, err)
func CheckEventInvitationToken(tokenStr string, eventId *bson.ObjectId) (*string, error) {
	eventIdHex := eventId.Hex()
	token, err := jwt.Parse(
		tokenStr,
		func(token *jwt.Token) (interface{}, error) {
			return []byte(settings.EVENT_INVITE_SECRET + eventIdHex), nil
		},
	)
	if err != nil {
		return nil, err
	}

	expectedAlg := settings.EVENT_INVITE_TOKEN_ALG
	tokenAlg := token.Header["alg"]
	if expectedAlg != tokenAlg {
		return nil, fmt.Errorf(
			"Expected %s signing method but token specified %s",
			expectedAlg,
			tokenAlg,
		)
	}

	if !token.Valid {
		return nil, fmt.Errorf("token.Valid == false")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid claims == %#v", claims)
	}
	tokenEmail := claims["email"]
	tokenEventIdHexIn := claims["eventId"]

	email, ok := tokenEmail.(string)
	if !ok {
		return nil, fmt.Errorf("tokenEmail == %#v", tokenEmail)
	}

	tokenEventIdHex, ok := tokenEventIdHexIn.(string)
	if !ok {
		return nil, fmt.Errorf("tokenEventIdHexIn == %#v", tokenEventIdHexIn)
	}

	if tokenEventIdHex != eventIdHex {
		return nil, fmt.Errorf(
			"MISMATCH eventId: %#v == %#v",
			tokenEventIdHex,
			eventIdHex,
		)
	}

	return &email, nil
}

func newEventInvitationToken(eventId bson.ObjectId, email string) (string, time.Time) {
	eventIdHex := eventId.Hex()
	now := time.Now()
	exp := now.Add(time.Duration(settings.EVENT_INVITE_TOKEN_EXP_SECONDS) * time.Second)
	tokenStr, _ := jwt.NewWithClaims(
		jwt.GetSigningMethod(settings.EVENT_INVITE_TOKEN_ALG),
		jwt.MapClaims{
			"eventId": eventIdHex,
			"email":   email,
			"iat":     now.Unix(),
			"exp":     exp.Unix(),
		},
	).SignedString([]byte(
		settings.EVENT_INVITE_SECRET + eventIdHex,
	))
	return tokenStr, exp
}
