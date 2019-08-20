package event_lib

import (
	"bytes"
	"errors"
	"fmt"
	"net/url"
	"scal"
	"scal/settings"
	"scal/storage"
	. "scal/user_lib"
	"text/template"
	"time"

	"github.com/globalsign/mgo/bson"
	"github.com/ilius/ripo"
	//"net/url"
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
	EventId   string `bson:"_id,objectid"`
	EventType string `bson:"eventType"`

	CreationTime time.Time            `bson:"creationTime"`
	FieldsMtime  map[string]time.Time `bson:"fieldsMtime"`

	OwnerEmail   string   `bson:"ownerEmail"`
	IsPublic     bool     `bson:"isPublic"`
	AccessEmails []string `bson:"accessEmails"`

	GroupId    *string          `bson:"groupId,objectid"`
	GroupModel *EventGroupModel `bson:"-"`

	// PublicJoinPolicy string         `bson:"publicJoinPolicy"` // not indexed
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

func (model EventMetaModel) UniqueM() scal.M {
	return scal.M{
		"_id": bson.ObjectIdHex(model.EventId),
	}
}

func (model EventMetaModel) Collection() string {
	return storage.C_eventMeta
}

func (model EventMetaModel) GroupIdHex() string {
	if model.GroupId != nil {
		return *model.GroupId
	}
	return ""
}

func (model EventMetaModel) JsonM() scal.M {
	return scal.M{
		"creationTime":   model.CreationTime,
		"ownerEmail":     model.OwnerEmail,
		"isPublic":       model.IsPublic,
		"accessEmails":   model.AccessEmails,
		"groupId":        model.GroupIdHex(),
		"publicJoinOpen": model.PublicJoinOpen,
		"maxAttendees":   model.MaxAttendees,
	}
}

func (model *EventMetaModel) CanReadFull(email string) bool {
	if email == model.OwnerEmail {
		return true
	}
	for _, aEmail := range model.AccessEmails {
		if email == aEmail {
			return true
		}
	}
	if model.GroupModel != nil {
		for _, aEmail := range model.GroupModel.ReadAccessEmails {
			if email == aEmail {
				return true
			}
		}
	}
	return false
}

func (model *EventMetaModel) CanRead(email string) bool {
	if model.IsPublic {
		return true
	}
	return model.CanReadFull(email)
}

func (model *EventMetaModel) GetAttending(
	db storage.Database,
	email string,
) string {
	// returns YES, NO, or MAYBE
	attendingModel, _ := LoadEventAttendingModel(db, model.EventId, email)
	return attendingModel.Attending
}

func (model *EventMetaModel) SetAttending(
	db storage.Database,
	email string,
	attending string,
) error {
	// attending: YES, NO, or MAYBE
	attendingModel, err := LoadEventAttendingModel(db, model.EventId, email)
	if err != nil {
		return err
	}
	attendingModel.Attending = attending
	attendingModel.ModifiedTime = time.Now()
	err = attendingModel.Save(db)
	return err
}

func (model *EventMetaModel) AttendingStatusCount(
	db storage.Database,
	attending string,
) (int, error) {
	return db.FindCount(
		storage.C_attending,
		scal.M{
			"eventId":   model.EventId,
			"attending": attending,
		},
	)
}

func (model *EventMetaModel) Join(db storage.Database, email string) error {
	if model.GetAttending(db, email) == YES {
		// does not make any changes on `model`
		if settings.ALLOW_REJOIN_EVENT {
			return nil
		}
		return errors.New("you have already joined this event")
	}
	if !model.CanReadFull(email) {
		if model.IsPublic {
			if !model.PublicJoinOpen {
				return errors.New("this public event is not open for joining")
			}
		} else {
			return errors.New("no access, no join")
		}
	}
	if model.MaxAttendees > 0 {
		attendingCount, err := model.AttendingStatusCount(db, YES)
		if err != nil {
			return err
		}
		if attendingCount >= model.MaxAttendees {
			return errors.New("maximum attendees exceeded, can not join event")
		}
	}
	model.SetAttending(db, email, YES)
	return nil
}

func (model *EventMetaModel) Leave(db storage.Database, email string) error {
	// does not make any changes on model
	if model.GetAttending(db, email) == NO {
		if model.CanReadFull(email) {
			return errors.New("you are not attending for this event")
		}
	}
	model.SetAttending(db, email, NO)
	return nil
}

func (model *EventMetaModel) PublicCanJoin() bool {
	return model.IsPublic && model.PublicJoinOpen
}

func (model *EventMetaModel) Invite(
	db storage.Database,
	email string,
	inviteEmails []string,
	remoteIp string,
	host string,
) error {
	var err error
	if len(inviteEmails) == 0 {
		return ripo.NewError(ripo.InvalidArgument, "missing 'inviteEmails'", nil)
	}

	now := time.Now()

	fullAc := model.CanReadFull(email)
	public := model.PublicCanJoin()
	if !(fullAc || public) {
		return ripo.NewError(ripo.PermissionDenied, "not allowed to invite to this event", nil)
	}
	if fullAc {
		accessEmailsMap := make(map[string]bool)
		newAccessEmails := make(
			[]string,
			0,
			len(model.AccessEmails)+len(inviteEmails),
		)
		for _, aEmail := range model.AccessEmails {
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
			EventId:  model.EventId,
			FuncName: "SetEventAccess",
			AccessEmails: &[2][]string{
				model.AccessEmails,
				newAccessEmails,
			},
		}
		model.AccessEmails = newAccessEmails
		err = db.Insert(metaChangeLog)
		if err != nil {
			return ripo.NewError(ripo.Internal, "", err)
		}
		err = db.Update(model)
		if err != nil {
			return ripo.NewError(ripo.Internal, "", err)
		}
	}

	eventRev, err := LoadLastRevisionModel(db, &model.EventId)
	if err != nil {
		if db.IsNotFound(err) {
			return ripo.NewError(ripo.NotFound, "event not found", err)
		}
		return ripo.NewError(ripo.Internal, "", err)
	}
	eventModel, err := LoadBaseEventModel(db, eventRev.Sha1)
	if err != nil {
		return ripo.NewError(ripo.Internal, "", err)
	}

	tplText := settings.EVENT_INVITE_EMAIL_TEMPLATE
	user := UserModelByEmail(email, db)
	if user == nil {
		return ripo.NewError(ripo.NotFound, "user not found", nil)
		// FIXME: or Internal?
	}
	for _, inviteEmail := range inviteEmails {
		subject := "Invitation to event: " + eventModel.Summary
		tpl, err := template.New(subject).Parse(tplText)
		if err != nil {
			return ripo.NewError(ripo.Internal, "", err)
		}
		var inviteName string
		inviteUser := UserModelByEmail(inviteEmail, db)
		if inviteUser == nil {
			// return errors.New("invited email not found"), scal.BadRequest
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
			EventType:   model.EventType,
			EventId:     model.EventId,
			Host:        host,
		}

		{
			token, _ := newEventInvitationToken(model.EventId, inviteEmail)
			tokenEscaped := url.QueryEscape(token) // Go < 1.8
			// tokenEscaped := url.PathEscape(token) // Go 1.8
			tplParams.TokenEscaped = tokenEscaped
		}

		buf := bytes.NewBufferString("")
		err = tpl.Execute(buf, tplParams)
		if err != nil {
			return ripo.NewError(ripo.Internal, "", err)
		}
		emailBody := buf.String()
		db.Insert(EventInvitationModel{
			Time:         now,
			SenderEmail:  email,
			InvitedEmail: inviteEmail,
			EventId:      model.EventId,
		})

		err = scal.SendEmail(&scal.SendEmailInput{
			To:      inviteEmail,
			Subject: subject,
			IsHtml:  false,
			Body:    emailBody,
		})
		if err != nil {
			fmt.Println("Failed to send email:", err)
			fmt.Println(emailBody)
		}
	}
	return nil
}

func (model *EventMetaModel) GetEmailsByAttendingStatus(
	db storage.Database,
	attending string,
) []string {
	emailStructs := []struct {
		Email string `bson:"email"`
	}{}
	cond := db.NewCondition(storage.AND)
	cond.Equals("eventId", model.EventId)
	cond.Equals("attending", attending)
	err := db.FindAll(&emailStructs, &storage.FindInput{
		Collection: storage.C_attending,
		Condition:  cond,
	})
	if err != nil {
		log.Error(fmt.Sprintf(
			"Internal Error: GetAttendingEmails: eventId=%v: %s\n",
			model.EventId,
			err.Error(),
		))
	}
	emails := make([]string, len(emailStructs))
	for i, m := range emailStructs {
		emails[i] = m.Email
	}
	return emails
}

func (model *EventMetaModel) GetAttendingEmails(db storage.Database) []string {
	return model.GetEmailsByAttendingStatus(db, YES)
}

func (model *EventMetaModel) GetNotAttendingEmails(db storage.Database) []string {
	return model.GetEmailsByAttendingStatus(db, NO)
}

func (model *EventMetaModel) GetMaybeAttendingEmails(db storage.Database) []string {
	return model.GetEmailsByAttendingStatus(db, MAYBE)
}

func LoadEventMetaModel(
	db storage.Database,
	eventId *string,
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
	Time    time.Time `bson:"time"`
	Email   string    `bson:"email"`
	EventId string    `bson:"eventId,objectid"`

	RemoteIp      string    `bson:"remoteIp"`
	TokenIssuedAt time.Time `bson:"tokenIssuedAt"`

	FuncName string `bson:"funcName"`

	GroupId        *[2]*string  `bson:"groupId,omitempty"`
	OwnerEmail     *[2]*string  `bson:"ownerEmail,omitempty"`
	IsPublic       *[2]bool     `bson:"isPublic,omitempty"`
	AccessEmails   *[2][]string `bson:"accessEmails,omitempty"`
	PublicJoinOpen *[2]bool     `bson:"publicJoinOpen,omitempty"`
	MaxAttendees   *[2]int      `bson:"maxAttendees,omitempty"`
}

func (model EventMetaChangeLogModel) Collection() string {
	return storage.C_eventMetaChangeLog
}

type EventInvitationModel struct {
	Time         time.Time `bson:"time"`
	SenderEmail  string    `bson:"senderEmail"`
	InvitedEmail string    `bson:"invitedEmail"`
	EventId      string    `bson:"eventId,objectid"`
}

func (model EventInvitationModel) Collection() string {
	return storage.C_invitation
}
