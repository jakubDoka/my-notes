package mongo

import (
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
func Or(targets ...bson.M) bson.M {
	return bson.M{"$or": targets}
}

// ID returns filter for searching by id
func ID(id core.ID) bson.M {
	return bson.M{"_id": id}
}

// NoteFilter creates filter for searching notes
func NoteFilter(values url.Values, published bool) bson.D {
	filter := bson.D{}

	if val := values["author"]; val[0] != "" {
		filter = append(filter, E("author", val))
	}

	if val := values["name"]; val[0] != "" {
		filter = append(filter, StartsWith("name", val[0]))
	}

	filter = append(filter, E("school", School(values["school"][0])))

	if val := values["year"]; val[0] != "" {
		i, err := strconv.Atoi(val[0])
		if err == nil {
			filter = append(filter, E("year", i))
		}
	}

	if val := values["month"]; val[0] != "" {
		i, err := strconv.Atoi(val[0])
		if err == nil {
			filter = append(filter, E("month", i))
		}
	}

	if val := values["subject"]; val[0] != "" {
		filter = append(filter, StartsWith("subject", val[0]))
	}

	if val := values["theme"]; val[0] != "" {
		filter = append(filter, E("theme", val[0]))
	}

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
	return map[string]int{
		"elementary-middle": 0,
		"high":              1,
		"university":        2,
	}[name]
}
