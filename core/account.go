package core

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/jakubDoka/sterr"
)

/*imp(
	gogen/templates
)*/

/*gen(
	templates.Vec<ID, IDS>
)*/

// ID is type alias for id datatype that should be used everywhere when
// refering to database id
type ID = uint64

// None is none id placeholder
const None = math.MaxUint64

// IDer is something that has id
type IDer interface {
	AID() ID
}

// BiSearch function for IDS binary search
var BiSearch = func(a, b ID) uint8 {
	if a == b {
		return 0
	} else if a > b {
		return 1
	}
	return 2
}

// ActionSpacing is period between actions to prevent spam
const ActionSpacing = time.Minute * 5

// Errors ...
var (
	ErrImpossible        = sterr.New("this error should not be possible under expected corcompstances, please report this")
	ErrInvalidTargetType = sterr.New("failed to parse target")
)

// EI is shorthand for ErrImpossible wrapper
func EI(err error) error {
	return ErrImpossible.Wrap(err)
}

// RawID is used for retrieving just ids
type RawID struct {
	ID ID `bson:"_id"`
}

// Account is user account, it contains password in form oh hash
type Account struct {
	ID ID `bson:"_id"`

	BornDate, LastAction int64

	Name, Password, Code, Email string

	Cfg Config
}

// AID implements IDer
func (a *Account) AID() ID { return a.ID }

// Cookie produces account cookie
func (a *Account) Cookie() http.Cookie {
	return http.Cookie{Name: "user", Value: a.Name + " " + a.Password}
}

// Censure censures all private information of user
func (a *Account) Censure() {
	a.Password = "i don't think so"
	a.Email = "not.quite@gamil.com"
}

// Config ...
type Config struct {
	Colors []string
}

// Note ...
type Note struct {
	ID     ID `bson:"_id"`
	Author ID

	Likes IDS

	Month, Year, School int

	BornDate int64

	Content string

	Published bool

	Theme, Subject, Name string

	Comments []ID
}

// AID implements IDer
func (a *Note) AID() ID { return a.ID }

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
	Likes    IDS
	Content  string
}

// As String for testing purposes
func (n *NotePreview) String() string {
	return fmt.Sprintf("%d, %d, %s, %v, %s", n.ID, n.Author, n.Name, n.Likes, n.Content)
}

// Comment ...
type Comment struct {
	ID           ID `bson:"_id"`
	BornDate     int64
	Likes        IDS
	Target       Target
	Author, Note ID
	Content      string
}

// AID implements IDer
func (a *Comment) AID() ID { return a.ID }

// TargetType ...
type TargetType uint8

// Target variants
const (
	NoteT TargetType = iota
	CommentT
)

// Target stores information about what the document is linked to
type Target struct {
	Type TargetType
	ID   ID `bson:"id"`
}

// Likes is ofr withdrawing only likes from a document
type Likes struct {
	Likes IDS
}

// ParseTargetType parses string to target type
func ParseTargetType(tp string) (TargetType, error) {
	t, ok := map[string]TargetType{
		"comment": CommentT,
		"note":    NoteT,
	}[tp]

	if !ok {
		return 0, ErrInvalidTargetType
	}

	return t, nil
}

// ParseID ...
func ParseID(id string) (ID, error) {
	i, err := strconv.ParseUint(id, 10, 64)
	return i, EI(err)
}

// Time returns current time in millis
func Time() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

// FormatTime formats time conveniently
func FormatTime(millis int64) string {
	for _, p := range []struct {
		d time.Duration
		s string
	}{
		{time.Millisecond, "millis"},
		{time.Second, "s"},
		{time.Minute, "m"},
		{time.Hour, "h"},
		{time.Hour * 24, "d"},
		{time.Hour * 24 * 356, "y"},
	} {
		if int64(p.d) >= millis*int64(time.Millisecond) {
			return fmt.Sprintf(p.s, millis/int64(p.d))
		}
	}

	return ""
}

// SearchRequest ...
type SearchRequest struct {
	Name, School, Theme, Author, Subject string
	Year, Month                          int
}
