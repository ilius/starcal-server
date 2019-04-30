package event_lib

import (
	"github.com/ilius/ripo"

	"github.com/globalsign/mgo/bson"

	"scal"
	"scal/storage"
)

type EventGroupModel struct {
	Id               bson.ObjectId `bson:"_id" json:"groupId"`
	Title            string        `bson:"title" json:"title"`
	OwnerEmail       string        `bson:"ownerEmail" json:"ownerEmail"`
	AddAccessEmails  []string      `bson:"addAccessEmails,omitempty" json:"addAccessEmails,omitempty"`
	ReadAccessEmails []string      `bson:"readAccessEmails,omitempty" json:"readAccessEmails,omitempty"`
}

func (model EventGroupModel) UniqueM() scal.M {
	return scal.M{
		"_id": model.Id,
	}
}
func (EventGroupModel) Collection() string {
	return storage.C_group
}
func (model EventGroupModel) EmailCanAdd(email string) bool {
	if email == model.OwnerEmail {
		return true
	}
	for _, aEmail := range model.AddAccessEmails {
		if email == aEmail {
			return true
		}
	}
	return false
}
func (model EventGroupModel) CanRead(email string) bool {
	if email == model.OwnerEmail {
		return true
	}
	for _, aEmail := range model.ReadAccessEmails {
		if email == aEmail {
			return true
		}
	}
	return false
}

func LoadGroupModelById(
	attrName string,
	db storage.Database,
	groupId *bson.ObjectId,
) (*EventGroupModel, error) {
	if groupId == nil {
		return nil, ripo.NewError(ripo.InvalidArgument, "missing '"+attrName+"'", nil)
	}
	groupModel := &EventGroupModel{
		Id: *groupId,
	}
	err := db.Get(groupModel)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, ripo.NewError(ripo.NotFound, "group not found", err)
		}
		return nil, ripo.NewError(ripo.Internal, "", err)
	}
	return groupModel, nil
}

func LoadGroupModelByIdHex(
	attrName string,
	db storage.Database,
	groupIdHex string,
) (*EventGroupModel, error) {
	if groupIdHex == "" {
		return nil, ripo.NewError(ripo.InvalidArgument, "missing '"+attrName+"'", nil)
	}
	if !bson.IsObjectIdHex(groupIdHex) { // to avoid panic!
		return nil, ripo.NewError(ripo.InvalidArgument, "invalid '"+attrName+"'", nil)
	}
	groupId := bson.ObjectIdHex(groupIdHex)
	return LoadGroupModelById(
		attrName,
		db,
		&groupId,
	)
}
