package mongo

import (
	"gogen/str"
	"myNotes/core"
	"net/url"
	"strconv"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// All is empty filter
var All = bson.M{}

// Set update shorthand
func Set(target bson.M) bson.M {
	return bson.M{"$set": target}
}

// Inc update shotrand
func Inc(target bson.M) bson.M {
	return bson.M{"$inc": target}
}

// Or takes bson filters and creates OR query
func Or(targets ...interface{}) bson.M {
	return bson.M{"$or": targets}
}

// ID returns filter for searching by id
func ID(id core.ID) bson.M {
	return bson.M{"_id": id}
}

// NoteFilter creates filter for searching notes, passed url values have to contain keys with non empty
// lists even if you are not filtering them, if first value under key is "" then its ignored
func (d *DB) NoteFilter(values url.Values, published bool) bson.D {
	filter := bson.D{}

	// author is really annoing but important
	if val := values["author"][0]; val != "" {
		if str.StartsWith(val, ExactLabel) { // take care of exact
			ac, err := d.AccountByName(val[len(ExactLabel):])
			if err == nil {
				filter = append(filter, E("author", ac.ID))
			}
		} else { // worst part, we have to collect ids of all possible authors
			ids, err := d.AccountIdsForName(val)
			if err != nil {
				panic(err)
			}
			if len(ids) != 0 {
				eIds := make([]interface{}, len(ids))
				for i, id := range ids {
					eIds[i] = bson.M{"author": id.ID}
				}
				filter = append(filter, E("$or", eIds))
			}
		}
	}

	// again if string query starts with ExactLabel we will pick only exact matches
	// othervise use start with operation
	for _, field := range []string{"subject", "theme", "name"} {
		if val := values[field][0]; val != "" {
			if str.StartsWith(val, ExactLabel) {
				filter = append(filter, E(field, val))
			} else {
				filter = append(filter, StartsWith(field, val[len(ExactLabel):]))
			}
		}
	}

	for _, field := range []string{"year", "month"} {
		if val := values[field][0]; val != "" {
			i, err := strconv.Atoi(val)
			if err == nil {
				filter = append(filter, E(field, i))
			}
		}
	}

	filter = append(filter, E("school", School(values["school"][0])))

	if published {
		filter = append(filter, E("published", true))
	}

	return filter
}

// StartsWith is query based of fields string start
func StartsWith(field, sub string) bson.E {
	return E(field, primitive.Regex{Pattern: sub})
}

// E to prevent vet
func E(key string, value interface{}) bson.E {
	return bson.E{Key: key, Value: value}
}

// School converts string to coresponding int value
func School(name string) int {
	name = strings.ToLower(name)
	// 0 is considered none and so whatewer is inputted that is not contained in map will be none
	return map[string]int{
		"elementary-middle": 1,
		"high":              2,
		"university":        3,
	}[name]
}
