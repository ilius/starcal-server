package api_v1

import (
	"scal"
	"scal/event_lib"
	"scal/storage"
	"scal/user_lib"
	"testing"
	"time"

	"github.com/globalsign/mgo/bson"
	"github.com/golang/mock/gomock"
	"github.com/ilius/is"
	"github.com/ilius/libgostarcal/utils"
	. "github.com/ilius/ripo"
)

func NewTestHelper(t *testing.T, userEmail string) *TestHelper {
	userAuth := ""
	if userEmail != "" {
		var err error
		userAuth, err = utils.GenerateRandomBase64String(16)
		if err != nil {
			panic(err)
		}
	}
	return &TestHelper{
		t:         t,
		userEmail: userEmail,
		userAuth:  userAuth,
	}
}

type TestHelper struct {
	t *testing.T // set in NewTestHelper

	userEmail      string // set in NewTestHelper
	userAuth       string // set in NewTestHelper
	emailConfirmed bool   // set in SetEmailConfirmed

	is           *is.Is                     // set on Start()
	db           storage.Database           // set on Start()
	userModel    *user_lib.UserModel        // set on Start()
	defaultGroup *event_lib.EventGroupModel // set on Start()
	startTime    *time.Time                 // set on Start()

	mockControllers []*gomock.Controller // added to in NewRequestMock
}

func (h *TestHelper) SetUserAuth(userAuth string) {
	h.userAuth = userAuth
}

func (h *TestHelper) Start() {
	var err error
	now := time.Now()
	h.startTime = &now
	h.is = is.New(h.t)
	h.db, err = storage.GetDB()
	if err != nil {
		panic(err)
	}
	if h.userEmail != "" {
		h.createUser()
	}
}

func (h *TestHelper) Is() *is.Is {
	return h.is
}

func (h *TestHelper) DB() storage.Database {
	return h.db
}

func (h *TestHelper) UserModel() *user_lib.UserModel {
	return h.userModel
}

func (h *TestHelper) SetEmailConfirmed(emailConfirmed bool) *TestHelper {
	h.emailConfirmed = emailConfirmed
	return h
}

func (h *TestHelper) createUser() {
	db := h.db
	email := h.userEmail
	now := time.Now()
	h.userModel = &user_lib.UserModel{
		Id:             bson.NewObjectId().Hex(),
		Email:          email,
		EmailConfirmed: h.emailConfirmed,
		TokenIssuedAt:  &now,
	}

	if h.userAuth != "" {
		if testUserMap == nil {
			testUserMap = map[string]*user_lib.UserModel{}
		}
		testUserMap[h.userAuth] = h.userModel
	}

	h.defaultGroup = &event_lib.EventGroupModel{
		Id:         bson.NewObjectId().Hex(),
		Title:      email,
		OwnerEmail: email,
	}
	err := db.Insert(h.defaultGroup)
	if err != nil {
		panic(err)
	}
	h.userModel.DefaultGroupId = &h.defaultGroup.Id

	err = db.Insert(h.userModel)
	if err != nil {
		panic(err)
	}
}

func (h *TestHelper) DefaultGroup() *event_lib.EventGroupModel {
	return h.defaultGroup
}

func (h *TestHelper) CreateGroup(title string) *event_lib.EventGroupModel {
	groupModel := &event_lib.EventGroupModel{}

	groupId := bson.NewObjectId().Hex()
	groupModel.Id = groupId
	groupModel.OwnerEmail = h.userEmail
	groupModel.Title = title
	// groupModel.AddAccessEmails = addAccessEmails
	// groupModel.ReadAccessEmails = readAccessEmails

	err := h.db.Insert(groupModel)
	if err != nil {
		panic(err)
	}
	return groupModel
}

func (h *TestHelper) NewRequestMock(authHeader bool, remoteIp string) (ExtendedRequest, *MockExtendedRequest) {
	ctrl := gomock.NewController(h.t)
	h.mockControllers = append(h.mockControllers, ctrl)
	mockReq := NewMockExtendedRequest(ctrl)
	if authHeader {
		mockReq.EXPECT().Header(gomock.Eq("Authorization")).Return(h.userAuth)
	}
	if remoteIp != "" {
		mockReq.EXPECT().RemoteIP().Return(remoteIp, nil)
	}
	return mockReq, mockReq
}

func (h *TestHelper) FinishMocks() {
	for _, ctrl := range h.mockControllers {
		ctrl.Finish()
	}
	h.mockControllers = nil
}

func (h *TestHelper) deleteUser() {
	if h.userModel == nil {
		return
	}
	db := h.db
	err := db.Remove(h.userModel)
	if err != nil {
		panic(err)
	}
	if h.defaultGroup != nil {
		err := db.Remove(h.defaultGroup)
		if err != nil {
			panic(err)
		}
	}
}

func (h *TestHelper) DeleteUserByEmail(email string) {
	userModel := &user_lib.UserModel{
		Email: email,
	}
	db := h.db
	err := db.Remove(userModel)
	if err != nil {
		panic(err)
	}
}

func (h *TestHelper) RemoveLoginHistory() {
	startTime := *h.startTime
	db := h.db
	count, err := db.RemoveAll(storage.C_userLogins, scal.M{
		"email": h.userEmail,
		"time": scal.M{
			"$gte": startTime,
		},
	})
	if err != nil {
		panic(err)
	}
	h.t.Logf("Removed %d login entries", count)
}

func (h *TestHelper) Finish() {
	testUserMapClear()
	h.deleteUser()
	h.FinishMocks()
}
