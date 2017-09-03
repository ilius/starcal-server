package event_lib

import (
	"github.com/ilius/restpc"

	"gopkg.in/mgo.v2/bson"

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

func (self EventGroupModel) UniqueM() scal.M {
	return scal.M{
		"_id": self.Id,
	}
}
func (self EventGroupModel) Collection() string {
	return storage.C_group
}
func (self EventGroupModel) EmailCanAdd(email string) bool {
	if email == self.OwnerEmail {
		return true
	}
	for _, aEmail := range self.AddAccessEmails {
		if email == aEmail {
			return true
		}
	}
	return false
}
func (self EventGroupModel) CanRead(email string) bool {
	if email == self.OwnerEmail {
		return true
	}
	for _, aEmail := range self.ReadAccessEmails {
		if email == aEmail {
			return true
		}
	}
	return false
}

func (self *EventGroupModel) GetAccessCond(email string) scal.M {
	if self.CanRead(email) {
		return scal.M{}
	} else {
		return scal.M{
			"$or": []scal.M{
				scal.M{"ownerEmail": email},
				scal.M{"isPublic": true},
				scal.M{"accessEmails": email},
			},
		}
	}
}
func (self *EventGroupModel) GetLookupMetaAccessPipeline(
	email string,
	localField string,
) []scal.M {
	if self.CanRead(email) {
		return []scal.M{}
	} else {
		return []scal.M{
			{"$lookup": scal.M{
				"from":         storage.C_eventMeta,
				"localField":   localField,
				"foreignField": "_id",
				"as":           "meta",
			}},
			{"$unwind": "$meta"},
			{"$match": scal.M{
				"$or": []scal.M{
					scal.M{"meta.ownerEmail": email},
					scal.M{"meta.isPublic": true},
					scal.M{"meta.accessEmails": email},
				},
			}},
		}
	}
}

func LoadGroupModelById(
	attrName string,
	db storage.Database,
	groupId *bson.ObjectId,
) (*EventGroupModel, error) {
	if groupId == nil {
		return nil, restpc.NewError(restpc.InvalidArgument, "missing '"+attrName+"'", nil)
	}
	groupModel := &EventGroupModel{
		Id: *groupId,
	}
	err := db.Get(groupModel)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, restpc.NewError(restpc.NotFound, "group not found", err)
		}
		return nil, restpc.NewError(restpc.Internal, "", err)
	}
	return groupModel, nil
}

func LoadGroupModelByIdHex(
	attrName string,
	db storage.Database,
	groupIdHex string,
) (*EventGroupModel, error) {
	if groupIdHex == "" {
		return nil, restpc.NewError(restpc.InvalidArgument, "missing '"+attrName+"'", nil)
	}
	if !bson.IsObjectIdHex(groupIdHex) { // to avoid panic!
		return nil, restpc.NewError(restpc.InvalidArgument, "invalid '"+attrName+"'", nil)
	}
	groupId := bson.ObjectIdHex(groupIdHex)
	return LoadGroupModelById(
		attrName,
		db,
		&groupId,
	)
}
