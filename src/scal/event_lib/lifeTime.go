package event_lib

import "gopkg.in/mgo.v2"
import "scal/storage"

type LifeTimeEventModel struct {
    BaseEventModel  `bson:",inline" json:",inline"`
    StartJd int     `bson:"startJd" json:"startJd"`
    EndJd int       `bson:"endJd" json:"endJd"`
}
func (self LifeTimeEventModel) Type() string {
    return "lifeTime"
}

func LoadLifeTimeEventModel(db *mgo.Database, sha1 string) (
    *LifeTimeEventModel,
    error,
) {
    model := LifeTimeEventModel{}
    model.Sha1 = sha1
    err := storage.Get(db, &model)
    return &model, err
}


type LifeTimeEvent struct {
    BaseEvent
    startJd int
    endJd int
}
func (self LifeTimeEvent) Type() string {
    return "lifeTime"
}
func (self LifeTimeEvent) StartJd() int {
    return self.startJd
}
func (self LifeTimeEvent) EndJd() int {
    return self.endJd
}


func (self LifeTimeEvent) Model() LifeTimeEventModel {
    return LifeTimeEventModel{
        BaseEventModel: self.BaseModel(),
        StartJd: self.startJd,
        EndJd: self.endJd,
    }
}
func (self LifeTimeEventModel) GetEvent() (LifeTimeEvent, error) {
    baseEvent, err := self.BaseEventModel.GetBaseEvent()
    if err != nil {
        return LifeTimeEvent{}, err
    }
    return LifeTimeEvent{
        BaseEvent: baseEvent,
        startJd: self.StartJd,
        endJd: self.EndJd,
    }, nil
}


