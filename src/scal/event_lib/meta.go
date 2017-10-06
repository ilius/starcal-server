package event_lib

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net/url"
	"text/template"
	"time"

	"github.com/ilius/restpc"
	//"net/url"

	"gopkg.in/mgo.v2/bson"

	"scal"
	"scal/settings"
	"scal/storage"
	. "scal/user_lib"
)

/*
// PublicJoinPolicy
const (
    FreeToJoin = "FreeToJoin"
    NoJoin = "NoJoin"
    //JoinNeedsVerify = "JoinNeedsVerify" // makes implementation more complex
)
*/

type EventMetaModel struct {
	EventId   bson.ObjectId `bson:"_id"`
	EventType string        `bson:"eventType"`

	CreationTime time.Time            `bson:"creationTime"`
	FieldsMtime  map[string]time.Time `bson:"fieldsMtime"`

	OwnerEmail   string   `bson:"ownerEmail"`
	IsPublic     bool     `bson:"isPublic"`
	AccessEmails []string `bson:"accessEmails"`

	GroupId    *bson.ObjectId   `bson:"groupId"`
	GroupModel *EventGroupModel `bson:"-"`

	//PublicJoinPolicy string         `bson:"publicJoinPolicy"` // not indexed
	PublicJoinOpen bool `bson:"publicJoinOpen"`
	MaxAttendees   int  `bson:"maxAttendees"`
}

type InviteEmailTemplateParams struct {
	EventModel   *BaseEventModel
	SenderEmail  string
	SenderName   string
	Email        string
	Name         string
	EventType    string
	EventId      string
	TokenEscaped string
	Host         string
}

func (self EventMetaModel) UniqueM() scal.M {
	return scal.M{
		"_id": self.EventId,
	}
}
func (self EventMetaModel) Collection() string {
	return storage.C_eventMeta
}
func (self EventMetaModel) GroupIdHex() string {
	if self.GroupId != nil {
		return self.GroupId.Hex()
	}
	return ""
}
func (self EventMetaModel) JsonM() scal.M {
	return scal.M{
		"creationTime":   self.CreationTime,
		"ownerEmail":     self.OwnerEmail,
		"isPublic":       self.IsPublic,
		"accessEmails":   self.AccessEmails,
		"groupId":        self.GroupIdHex(),
		"publicJoinOpen": self.PublicJoinOpen,
		"maxAttendees":   self.MaxAttendees,
	}
}
func (self *EventMetaModel) CanReadFull(email string) bool {
	if email == self.OwnerEmail {
		return true
	}
	for _, aEmail := range self.AccessEmails {
		if email == aEmail {
			return true
		}
	}
	if self.GroupModel != nil {
		for _, aEmail := range self.GroupModel.ReadAccessEmails {
			if email == aEmail {
				return true
			}
		}
	}
	return false
}
func (self *EventMetaModel) CanRead(email string) bool {
	if self.IsPublic {
		return true
	}
	return self.CanReadFull(email)
}
func (self *EventMetaModel) GetAttending(
	db storage.Database,
	email string,
) string {
	// returns YES, NO, or MAYBE
	attendingModel, _ := LoadEventAttendingModel(db, self.EventId, email)
	return attendingModel.Attending
}
func (self *EventMetaModel) SetAttending(
	db storage.Database,
	email string,
	attending string,
) error {
	// attending: YES, NO, or MAYBE
	attendingModel, err := LoadEventAttendingModel(db, self.EventId, email)
	if err != nil {
		return err
	}
	attendingModel.Attending = attending
	attendingModel.ModifiedTime = time.Now()
	err = attendingModel.Save(db)
	return err
}
func (self *EventMetaModel) AttendingStatusCount(
	db storage.Database,
	attending string,
) (int, error) {
	return db.FindCount(
		storage.C_attending,
		scal.M{
			"eventId":   self.EventId,
			"attending": attending,
		},
	)
}
func (self *EventMetaModel) Join(db storage.Database, email string) error {
	// does not make any changes on self
	if self.GetAttending(db, email) == YES {
		return errors.New("you have already joined this event")
	}
	if !self.CanReadFull(email) {
		if self.IsPublic {
			if !self.PublicJoinOpen {
				return errors.New("this public event is not open for joining")
			}
		} else {
			return errors.New("no access, no join")
		}
	}
	if self.MaxAttendees > 0 {
		attendingCount, err := self.AttendingStatusCount(db, YES)
		if err != nil {
			return err
		}
		if attendingCount >= self.MaxAttendees {
			return errors.New("maximum attendees exceeded, can not join event")
		}
	}
	self.SetAttending(db, email, YES)
	return nil
}
func (self *EventMetaModel) Leave(db storage.Database, email string) error {
	// does not make any changes on self
	if self.GetAttending(db, email) == NO {
		if self.CanReadFull(email) {
			return errors.New("you are not attending for this event")
		}
	}
	self.SetAttending(db, email, NO)
	return nil
}
func (self *EventMetaModel) PublicCanJoin() bool {
	return self.IsPublic && self.PublicJoinOpen
}
func (self *EventMetaModel) Invite(
	db storage.Database,
	email string,
	inviteEmails []string,
	remoteIp string,
	host string,
) error {
	var err error
	if len(inviteEmails) == 0 {
		return restpc.NewError(restpc.InvalidArgument, "missing 'inviteEmails'", nil)
	}

	now := time.Now()

	fullAc := self.CanReadFull(email)
	public := self.PublicCanJoin()
	if !(fullAc || public) {
		return restpc.NewError(restpc.PermissionDenied, "not allowed to invite to this event", nil)
	}
	if fullAc {
		accessEmailsMap := make(map[string]bool)
		newAccessEmails := make(
			[]string,
			0,
			len(self.AccessEmails)+len(inviteEmails),
		)
		for _, aEmail := range self.AccessEmails {
			accessEmailsMap[aEmail] = true
			newAccessEmails = append(newAccessEmails, aEmail)
		}
		for _, inviteEmail := range inviteEmails {
			_, found := accessEmailsMap[inviteEmail]
			if !found {
				newAccessEmails = append(newAccessEmails, inviteEmail)
			}
		}
		metaChangeLog := EventMetaChangeLogModel{
			Time:     now,
			Email:    email,
			RemoteIp: remoteIp,
			EventId:  self.EventId,
			FuncName: "SetEventAccess",
			AccessEmails: &[2][]string{
				self.AccessEmails,
				newAccessEmails,
			},
		}
		self.AccessEmails = newAccessEmails
		err = db.Insert(metaChangeLog)
		if err != nil {
			return restpc.NewError(restpc.Internal, "", err)
		}
		err = db.Update(self)
		if err != nil {
			return restpc.NewError(restpc.Internal, "", err)
		}
	}

	eventRev, err := LoadLastRevisionModel(db, &self.EventId)
	if err != nil {
		if db.IsNotFound(err) {
			return restpc.NewError(restpc.NotFound, "event not found", err)
		}
		return restpc.NewError(restpc.Internal, "", err)
	}
	eventModel, err := LoadBaseEventModel(db, eventRev.Sha1)
	if err != nil {
		return restpc.NewError(restpc.Internal, "", err)
	}

	tplText := settings.EVENT_INVITE_EMAIL_TEMPLATE
	user := UserModelByEmail(email, db)
	if user == nil {
		return restpc.NewError(restpc.NotFound, "user not found", nil)
		// FIXME: or Internal?
	}
	for _, inviteEmail := range inviteEmails {
		subject := "Invitation to event: " + eventModel.Summary
		tpl, err := template.New(subject).Parse(tplText)
		if err != nil {
			return restpc.NewError(restpc.Internal, "", err)
		}
		var inviteName string
		inviteUser := UserModelByEmail(inviteEmail, db)
		if inviteUser == nil {
			//return errors.New("invited email not found"), scal.BadRequest
			// FIXME
			inviteName = ""
		} else {
			inviteName = inviteUser.FullName
		}

		tplParams := InviteEmailTemplateParams{
			EventModel:  eventModel,
			SenderEmail: email,
			SenderName:  user.FullName,
			Email:       inviteEmail,
			Name:        inviteName,
			EventType:   self.EventType,
			EventId:     self.EventId.Hex(),
			Host:        host,
		}

		{
			token, _ := newEventInvitationToken(eventModel.Id, inviteEmail)
			tokenEscaped := url.QueryEscape(token) // Go < 1.8
			// tokenEscaped := url.PathEscape(token) // Go 1.8
			tplParams.TokenEscaped = tokenEscaped
		}

		buf := bytes.NewBufferString("")
		err = tpl.Execute(buf, tplParams)
		if err != nil {
			return restpc.NewError(restpc.Internal, "", err)
		}
		emailBody := buf.String()
		db.Insert(EventInvitationModel{
			Time:         now,
			SenderEmail:  email,
			InvitedEmail: inviteEmail,
			EventId:      self.EventId,
		})

		err = scal.SendEmail(
			inviteEmail,
			subject,
			false, // isHtml
			emailBody,
		)
		if err != nil {
			fmt.Println("Failed to send email:", err)
			fmt.Println(emailBody)
		}
	}
	return nil
}
func (self *EventMetaModel) GetEmailsByAttendingStatus(
	db storage.Database,
	attending string,
) []string {
	emailStructs := []struct {
		Email string `bson:"email"`
	}{}
	err := db.FindAll(
		storage.C_attending,
		scal.M{
			"eventId":   self.EventId,
			"attending": attending,
		},
		&emailStructs,
	)
	if err != nil {
		log.Printf(
			"Internal Error: GetAttendingEmails: eventId=%v: %s\n",
			self.EventId,
			err.Error(),
		)
	}
	emails := make([]string, len(emailStructs))
	for i, m := range emailStructs {
		emails[i] = m.Email
	}
	return emails
}
func (self *EventMetaModel) GetAttendingEmails(db storage.Database) []string {
	return self.GetEmailsByAttendingStatus(db, YES)
}
func (self *EventMetaModel) GetNotAttendingEmails(db storage.Database) []string {
	return self.GetEmailsByAttendingStatus(db, NO)
}
func (self *EventMetaModel) GetMaybeAttendingEmails(db storage.Database) []string {
	return self.GetEmailsByAttendingStatus(db, MAYBE)
}

func LoadEventMetaModel(
	db storage.Database,
	eventId *bson.ObjectId,
	loadGroup bool,
) (*EventMetaModel, error) {
	var err error
	eventMeta := EventMetaModel{
		EventId: *eventId,
	}
	err = db.Get(&eventMeta)
	if err != nil {
		return nil, err
	}
	if loadGroup && eventMeta.GroupId != nil {
		// groupModel, err, internalErr
		groupModel, err := LoadGroupModelById(
			"groupId",
			db,
			eventMeta.GroupId,
		)
		if err != nil {
			return nil, err
		}
		eventMeta.GroupModel = groupModel
	}
	return &eventMeta, nil
}

type EventMetaChangeLogModel struct {
	Time     time.Time     `bson:"time"`
	Email    string        `bson:"email"`
	RemoteIp string        `bson:"remoteIp"`
	EventId  bson.ObjectId `bson:"eventId"`
	FuncName string        `bson:"funcName"`

	GroupId        *[2]*bson.ObjectId `bson:"groupId,omitempty"`
	OwnerEmail     *[2]*string        `bson:"ownerEmail,omitempty"`
	IsPublic       *[2]bool           `bson:"isPublic,omitempty"`
	AccessEmails   *[2][]string       `bson:"accessEmails,omitempty"`
	PublicJoinOpen *[2]bool           `bson:"publicJoinOpen,omitempty"`
	MaxAttendees   *[2]int            `bson:"maxAttendees,omitempty"`
}

func (model EventMetaChangeLogModel) Collection() string {
	return storage.C_eventMetaChangeLog
}

type EventInvitationModel struct {
	Time         time.Time     `bson:"time"`
	SenderEmail  string        `bson:"senderEmail"`
	InvitedEmail string        `bson:"invitedEmail"`
	EventId      bson.ObjectId `bson:"eventId"`
}

func (model EventInvitationModel) Collection() string {
	return storage.C_invitation
}

func GetEventMetaPipeResults(
	db storage.Database,
	pipeline *[]scal.M,
) (*[]scal.M, error) {
	results := []scal.M{}
	for res := range db.PipeIter(storage.C_eventMeta, pipeline) {
		if err := res.Err; err != nil {
			return nil, err
		}
		if eventId, ok := res.M["_id"]; ok {
			res.M["eventId"] = eventId
			delete(res.M, "_id")
		}
		if dataI, ok := res.M["data"]; ok {
			data := dataI.(scal.M)
			delete(data, "_id")
			res.M["data"] = data
		}
		results = append(results, res.M)
	}
	return &results, nil
}
