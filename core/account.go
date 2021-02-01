package core

import (
	"net/http"
	"strconv"
)

// ID is type alias for id datatype that should be used everywhere when
// refering to database id
type ID = uint64

// Errors ...
var (
	ErrImpossible = NErr("error that just happened should not be possible under expected corcompstances, report this if you can reason about your actions")
)

// RawID is used for retrieving just ids
type RawID struct {
	ID ID `bson:"_id"`
}

// Account is user account, it contains password in form oh hash
type Account struct {
	ID    ID `bson:"_id"`
	Notes []ID

	Name, Password, Code, Email string

	Cfg Config
}

//#FF00FF #FF00FF #FF00FF

// Cookie produces account cookie
func (a *Account) Cookie() http.Cookie {
	return http.Cookie{Name: "user", Value: a.Name + " " + a.Password}
}

// Config ...
type Config struct {
	Colors []string
}

// Note ...
type Note struct {
	ID     ID `bson:"_id"`
	Author ID

	Likes, Month, Year, School int

	BornDate int64

	Content string

	Published bool

	Theme, Subject, Name string
}

// Draft ...
type Draft struct {
	ID                   ID `bson:"_id"`
	Month, Year          int
	Theme, Subject, Name string
	Published            bool
}

// NotePreview ...
type NotePreview struct {
	ID       ID `bson:"_id"`
	Author   ID
	BornDate uint64
	Name     string
	Likes    int
	Content  string
}

// ParseID ...
func ParseID(id string) (ID, error) {
	return strconv.ParseUint(id, 10, 64)
}
