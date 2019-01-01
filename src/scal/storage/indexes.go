package storage

import (
	"log"
	"scal/settings"
	"time"

	mgo "github.com/globalsign/mgo"
	"github.com/ilius/ripo"
)

func EnsureIndexes() {
	var db *MongoDatabase
	{
		dbI, err := GetDB()
		if err != nil {
			panic(err)
		}
		db = dbI.(*MongoDatabase)
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
	db.EnsureIndex(C_user, mgo.Index{
		Key:        []string{"email"},
		Unique:     true,
		DropDups:   false,
		Background: false,
		Sparse:     false,
	})

	db.EnsureIndex(C_userChangeLog, mgo.Index{
		Key:        []string{"email"},
		Unique:     false,
		DropDups:   false,
		Background: false,
		Sparse:     false,
	})

	db.EnsureIndex(C_userLogins, mgo.Index{
		Key:        []string{"email"},
		Unique:     false,
		DropDups:   false,
		Background: false,
		Sparse:     false,
	})

	db.EnsureIndex(C_userLogins, mgo.Index{
		Key:        []string{"time"},
		Unique:     false,
		DropDups:   false,
		Background: false,
		Sparse:     false,
	})

	db.EnsureIndex(C_group, mgo.Index{
		Key:        []string{"ownerEmail"},
		Unique:     false,
		DropDups:   false,
		Background: false,
		Sparse:     true,
	})

	db.EnsureIndex(C_group, mgo.Index{
		Key:        []string{"readAccessEmails"},
		Unique:     false,
		DropDups:   false,
		Background: false,
		Sparse:     true,
	})

	db.EnsureIndex(C_eventMeta, mgo.Index{
		Key:        []string{"ownerEmail"},
		Unique:     false,
		DropDups:   false,
		Background: false,
		Sparse:     false,
	})

	db.EnsureIndex(C_eventMeta, mgo.Index{
		Key:        []string{"groupId"},
		Unique:     false,
		DropDups:   false,
		Background: false,
		Sparse:     false,
	})

	db.EnsureIndex(C_eventMeta, mgo.Index{
		Key:        []string{"creationTime"},
		Unique:     false,
		DropDups:   false,
		Background: false,
		Sparse:     false,
	})

	db.EnsureIndex(C_attending, mgo.Index{
		Key:        []string{"eventId", "email"},
		Unique:     true,
		DropDups:   false,
		Background: false,
		Sparse:     false,
	})

	db.EnsureIndex(C_attending, mgo.Index{
		Key:        []string{"eventId", "attending"},
		Unique:     false,
		DropDups:   false,
		Background: false,
		Sparse:     false,
	})

	db.EnsureIndex(C_attending, mgo.Index{
		Key:        []string{"eventId"},
		Unique:     false,
		DropDups:   false,
		Background: false,
		Sparse:     false,
	})

	// db.EnsureIndex(C_eventMetaChangeLog, mgo.Index{
	// 	Key:        []string{"time"},
	// 	Unique:     false,
	// 	DropDups:   false,
	// 	Background: false,
	// 	Sparse:     false,
	// })
	// db.EnsureIndex(C_eventMetaChangeLog, mgo.Index{
	// 	Key:        []string{"email"},
	// 	Unique:     false,
	// 	DropDups:   false,
	// 	Background: false,
	// 	Sparse:     false,
	// })
	// db.EnsureIndex(C_eventMetaChangeLog, mgo.Index{
	// 	Key:        []string{"eventId"},
	// 	Unique:     false,
	// 	DropDups:   false,
	// 	Background: false,
	// 	Sparse:     false,
	// })
	// db.EnsureIndex(C_eventMetaChangeLog, mgo.Index{
	// 	Key:        []string{"eventType"},
	// 	Unique:     false,
	// 	DropDups:   false,
	// 	Background: false,
	// 	Sparse:     false,
	// })
	// db.EnsureIndex(C_eventMetaChangeLog, mgo.Index{
	// 	Key:        []string{"ownerEmail"},
	// 	Unique:     false,
	// 	DropDups:   false,
	// 	Background: false,
	// 	Sparse:     false,
	// })
	// db.EnsureIndex(C_eventMetaChangeLog, mgo.Index{
	// 	Key:        []string{"groupId"},
	// 	Unique:     false,
	// 	DropDups:   false,
	// 	Background: false,
	// 	Sparse:     false,
	// })
	// db.EnsureIndex(C_eventMetaChangeLog, mgo.Index{
	// 	Key:        []string{"accessEmails"},
	// 	Unique:     false,
	// 	DropDups:   false,
	// 	Background: false,
	// 	Sparse:     false,
	// })

	db.EnsureIndex(C_revision, mgo.Index{
		Key:        []string{"sha1"},
		Unique:     false,
		DropDups:   false,
		Background: false,
		Sparse:     false,
	})

	db.EnsureIndex(C_revision, mgo.Index{
		Key:        []string{"eventId"},
		Unique:     false,
		DropDups:   false,
		Background: false,
		Sparse:     false,
	})

	db.EnsureIndex(C_revision, mgo.Index{
		Key:        []string{"time"},
		Unique:     false,
		DropDups:   false,
		Background: false,
		Sparse:     false,
	})

	db.EnsureIndex(C_eventData, mgo.Index{
		Key:        []string{"sha1"},
		Unique:     true,
		DropDups:   false,
		Background: false,
		Sparse:     false,
	})

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

	db.EnsureIndex(C_resetPwToken, mgo.Index{
		Key:        []string{"token"},
		Unique:     true,
		DropDups:   false,
		Background: false,
		Sparse:     false,
	})

	db.EnsureIndex(C_resetPwToken, mgo.Index{
		Key:        []string{"email"},
		Unique:     false,
		DropDups:   false,
		Background: false,
		Sparse:     false,
	})

	db.EnsureIndexWithTTL(C_resetPwToken, mgo.Index{
		Key:         []string{"-issueTime"},
		Unique:      false,
		DropDups:    false,
		Background:  false,
		Sparse:      false,
		ExpireAfter: time.Second * settings.RESET_PASSWORD_EXP_SECONDS,
	})

	for codeStr, seconds := range settings.ERRORS_EXPIRE_AFTER_SECONDS {
		_, ok := ripo.ErrorCodeByName[codeStr]
		if !ok {
			log.Printf("invalid error code %#v in settings.ERRORS_EXPIRE_AFTER_SECONDS", codeStr)
			continue
		}
		db.EnsureIndexWithTTL(C_errorsPrefix+codeStr, mgo.Index{
			Key:         []string{"-time"},
			Unique:      false,
			DropDups:    false,
			Background:  false,
			Sparse:      false,
			ExpireAfter: time.Second * time.Duration(seconds),
		})
	}
}
