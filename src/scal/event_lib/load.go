// Do not modify this file, it's auto-generated
package event_lib

import "scal/storage"

func LoadAllDayTaskEventModel(db storage.Database, sha1 string) (
	*AllDayTaskEventModel,
	error,
) {
	model := AllDayTaskEventModel{}
	model.Sha1 = sha1
	err := db.Get(&model)
	return &model, err
}

func LoadCustomEventModel(db storage.Database, sha1 string) (
	*CustomEventModel,
	error,
) {
	model := CustomEventModel{}
	model.Sha1 = sha1
	err := db.Get(&model)
	return &model, err
}

func LoadDailyNoteEventModel(db storage.Database, sha1 string) (
	*DailyNoteEventModel,
	error,
) {
	model := DailyNoteEventModel{}
	model.Sha1 = sha1
	err := db.Get(&model)
	return &model, err
}

func LoadLargeScaleEventModel(db storage.Database, sha1 string) (
	*LargeScaleEventModel,
	error,
) {
	model := LargeScaleEventModel{}
	model.Sha1 = sha1
	err := db.Get(&model)
	return &model, err
}

func LoadLifeTimeEventModel(db storage.Database, sha1 string) (
	*LifeTimeEventModel,
	error,
) {
	model := LifeTimeEventModel{}
	model.Sha1 = sha1
	err := db.Get(&model)
	return &model, err
}

func LoadMonthlyEventModel(db storage.Database, sha1 string) (
	*MonthlyEventModel,
	error,
) {
	model := MonthlyEventModel{}
	model.Sha1 = sha1
	err := db.Get(&model)
	return &model, err
}

func LoadTaskEventModel(db storage.Database, sha1 string) (
	*TaskEventModel,
	error,
) {
	model := TaskEventModel{}
	model.Sha1 = sha1
	err := db.Get(&model)
	return &model, err
}

func LoadUniversityClassEventModel(db storage.Database, sha1 string) (
	*UniversityClassEventModel,
	error,
) {
	model := UniversityClassEventModel{}
	model.Sha1 = sha1
	err := db.Get(&model)
	return &model, err
}

func LoadUniversityExamEventModel(db storage.Database, sha1 string) (
	*UniversityExamEventModel,
	error,
) {
	model := UniversityExamEventModel{}
	model.Sha1 = sha1
	err := db.Get(&model)
	return &model, err
}

func LoadWeeklyEventModel(db storage.Database, sha1 string) (
	*WeeklyEventModel,
	error,
) {
	model := WeeklyEventModel{}
	model.Sha1 = sha1
	err := db.Get(&model)
	return &model, err
}

func LoadYearlyEventModel(db storage.Database, sha1 string) (
	*YearlyEventModel,
	error,
) {
	model := YearlyEventModel{}
	model.Sha1 = sha1
	err := db.Get(&model)
	return &model, err
}
