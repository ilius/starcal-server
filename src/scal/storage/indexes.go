package storage

import (
	"errors"
	"fmt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"scal/settings"
	"strings"
	"time"
)

func EnsureIndexes() {
	dbI, err := GetDB()
	if err != nil {
		panic(err)
	}
	db, ok := dbI.(*MongoDatabase)
	if !ok {
		panic(errors.New("could not de-interface Database object"))
	}
	/*
		With DropDups set to true, documents with the
		same key as a previously indexed one will be dropped rather than an
		error returned.

		If Background is true, other connections will be allowed to proceed
		using the collection without the index while it's being built. Note that
		the session executing EnsureIndex will be blocked for as long as it
		takes for the index to be built.

		If Sparse is true, only documents containing the provided Key fields
		will be included in the index. When using a sparse index for sorting,
		only indexed documents will be returned.
	*/
	err = db.C(C_user).EnsureIndex(mgo.Index{
		Key:        []string{"email"},
		Unique:     true,
		DropDups:   false,
		Background: false,
		Sparse:     false,
	})
	if err != nil {
		panic(err)
	}

	err = db.C(C_group).EnsureIndex(mgo.Index{
		Key:        []string{"ownerEmail"},
		Unique:     false,
		DropDups:   false,
		Background: false,
		Sparse:     true,
	})
	if err != nil {
		panic(err)
	}
	err = db.C(C_group).EnsureIndex(mgo.Index{
		Key:        []string{"readAccessEmails"},
		Unique:     false,
		DropDups:   false,
		Background: false,
		Sparse:     true,
	})
	if err != nil {
		panic(err)
	}
	err = db.C(C_eventMeta).EnsureIndex(mgo.Index{
		Key:        []string{"ownerEmail"},
		Unique:     false,
		DropDups:   false,
		Background: false,
		Sparse:     false,
	})
	if err != nil {
		panic(err)
	}
	err = db.C(C_eventMeta).EnsureIndex(mgo.Index{
		Key:        []string{"groupId"},
		Unique:     false,
		DropDups:   false,
		Background: false,
		Sparse:     false,
	})
	if err != nil {
		panic(err)
	}
	err = db.C(C_eventMeta).EnsureIndex(mgo.Index{
		Key:        []string{"creationTime"},
		Unique:     false,
		DropDups:   false,
		Background: false,
		Sparse:     false,
	})
	if err != nil {
		panic(err)
	}
	err = db.C(C_attending).EnsureIndex(mgo.Index{
		Key:        []string{"eventId", "email"},
		Unique:     true,
		DropDups:   false,
		Background: false,
		Sparse:     false,
	})
	if err != nil {
		panic(err)
	}
	err = db.C(C_attending).EnsureIndex(mgo.Index{
		Key:        []string{"eventId", "attending"},
		Unique:     false,
		DropDups:   false,
		Background: false,
		Sparse:     false,
	})
	if err != nil {
		panic(err)
	}
	err = db.C(C_attending).EnsureIndex(mgo.Index{
		Key:        []string{"eventId"},
		Unique:     false,
		DropDups:   false,
		Background: false,
		Sparse:     false,
	})
	if err != nil {
		panic(err)
	}
	/*db.C(C_eventMetaChangeLog).EnsureIndex(mgo.Index{
		Key: []string{"time"},
		Unique: false,
		DropDups: false,
		Background: false,
		Sparse: false,
	})
	if err != nil {
		panic(err)
	}*/
	/*db.C(C_eventMetaChangeLog).EnsureIndex(mgo.Index{
		Key: []string{"email"},
		Unique: false,
		DropDups: false,
		Background: false,
		Sparse: false,
	})
	if err != nil {
		panic(err)
	}*/
	/*db.C(C_eventMetaChangeLog).EnsureIndex(mgo.Index{
		Key: []string{"eventId"},
		Unique: false,
		DropDups: false,
		Background: false,
		Sparse: false,
	})
	if err != nil {
		panic(err)
	}*/
	/*db.C(C_eventMetaChangeLog).EnsureIndex(mgo.Index{
		Key: []string{"eventType"},
		Unique: false,
		DropDups: false,
		Background: false,
		Sparse: false,
	})
	if err != nil {
		panic(err)
	}*/
	/*db.C(C_eventMetaChangeLog).EnsureIndex(mgo.Index{
		Key: []string{"ownerEmail"},
		Unique: false,
		DropDups: false,
		Background: false,
		Sparse: false,
	})
	if err != nil {
		panic(err)
	}*/
	/*db.C(C_eventMetaChangeLog).EnsureIndex(mgo.Index{
		Key: []string{"groupId"},
		Unique: false,
		DropDups: false,
		Background: false,
		Sparse: false,
	})
	if err != nil {
		panic(err)
	}*/
	/*db.C(C_eventMetaChangeLog).EnsureIndex(mgo.Index{
		Key: []string{"accessEmails"},
		Unique: false,
		DropDups: false,
		Background: false,
		Sparse: false,
	})
	if err != nil {
		panic(err)
	}*/
	err = db.C(C_revision).EnsureIndex(mgo.Index{
		Key:        []string{"sha1"},
		Unique:     false,
		DropDups:   false,
		Background: false,
		Sparse:     false,
	})
	if err != nil {
		panic(err)
	}
	err = db.C(C_revision).EnsureIndex(mgo.Index{
		Key:        []string{"eventId"},
		Unique:     false,
		DropDups:   false,
		Background: false,
		Sparse:     false,
	})
	if err != nil {
		panic(err)
	}
	err = db.C(C_revision).EnsureIndex(mgo.Index{
		Key:        []string{"time"},
		Unique:     false,
		DropDups:   false,
		Background: false,
		Sparse:     false,
	})
	if err != nil {
		panic(err)
	}

	err = db.C(C_eventData).EnsureIndex(mgo.Index{
		Key:        []string{"sha1"},
		Unique:     true,
		DropDups:   false,
		Background: false,
		Sparse:     false,
	})
	if err != nil {
		panic(err)
	}
	/*
		for _, colName := range []string{
			"events_allDayTask",
			"events_custom",
			"events_dailyNote",
			"events_largeScale",
			"events_lifeTime",
			"events_monthly",
			"events_task",
			"events_universityClass",
			"events_universityExam",
			"events_weekly",
			"events_yearly",
		} {
			err = db.C(colName).EnsureIndex(mgo.Index{
				Key: []string{"sha1"},
				Unique: true,
				DropDups: false,
				Background: false,
				Sparse: false,
			})
			if err != nil {
				panic(err)
			}
		}
	*/

	db.C(C_resetPwToken).EnsureIndex(mgo.Index{
		Key:        []string{"token"},
		Unique:     true,
		DropDups:   false,
		Background: false,
		Sparse:     false,
	})
	db.C(C_resetPwToken).EnsureIndex(mgo.Index{
		Key:        []string{"email"},
		Unique:     false,
		DropDups:   false,
		Background: false,
		Sparse:     false,
	})

	err = db.C(C_resetPwToken).EnsureIndex(mgo.Index{
		Key:         []string{"-issueTime"},
		Unique:      false,
		DropDups:    false,
		Background:  false,
		Sparse:      false,
		ExpireAfter: time.Second * settings.RESET_PASSWORD_EXP_SECONDS,
	})
	if err != nil {
		// if settings.RESET_PASSWORD_EXP_SECONDS is changed, we need to drop
		// and re-create the index, unless we use `collMod` added in 2.3.2
		// https://jira.mongodb.org/browse/SERVER-6700
		if strings.Contains(
			err.Error(),
			"already exists with different options",
		) {
			fmt.Printf(
				"Updating expireAfterSeconds on collection '%s'\n",
				C_resetPwToken,
			)
			err = db.Run(bson.D{
				{"collMod", C_resetPwToken},
				{"index", bson.M{
					"keyPattern":         bson.M{"issueTime": -1},
					"expireAfterSeconds": settings.RESET_PASSWORD_EXP_SECONDS,
				}},
			}, nil)
			if err != nil {
				panic(err)
			}
		} else {
			panic(err)
		}
	}
}
