package api_v1

import (
	"scal"
	"scal/storage"
	"fmt"
)

func GetEventMetaPipeResults(
	db storage.Database,
	pipelines *MongoPipelines,
	metaKeys []string,
) ([]scal.M, error) {
	results := []scal.M{}
	next, close := pipelines.Iter()
	defer close()
	for {
		row := scal.M{}
		err := next(&row)
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return nil, err
		}
		if eventId, ok := row["_id"]; ok {
			row["eventId"] = eventId
			delete(row, "_id")
		}
		if dataI, ok := row["data"]; ok {
			data, isM := dataI.(scal.M)
			if isM {
				delete(data, "_id")
				row["data"] = data
			} else {
				log.Error(fmt.Sprintf("type(dataI)=%T, dataI=%v\n", dataI, dataI))
			}
		}
		if metaKeys != nil {
			meta := scal.M{}
			for _, key := range metaKeys {
				if value, ok := row[key]; ok {
					meta[key] = value
					delete(row, key)
				}
			}
			row["meta"] = meta
		}
		results = append(results, row)
	}
	return results, nil
}
