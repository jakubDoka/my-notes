package mongo

import (
	"context"
	"math/rand"
	"myNotes/core"
	"net/url"
	"strconv"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Collection names
const (
	Accounts = "Accounts"
	Notes    = "Notes"
	CounterN = "CounterN"
	CounterA = "CounterA"

	Verified   = "ok"
	ExactLabel = "!"

	MaxCursorSize  = 50
	MaxPreviewSize = 400
)

// indexes
var (
	AccountIndex = []string{
		"name",
		"email",
	}

	NoteIndex = []string{
		"name",
		"school",
		"year",
		"month",
		"subject",
		"theme",
		"author",
	}
)

// MakeIndex creates indexing from list of field names
func MakeIndex(indexes []string) []mongo.IndexModel {
	idx := make([]mongo.IndexModel, len(indexes))
	for i, v := range indexes {
		idx[i].Keys = bson.M{v: 1}
	}
	return idx
}

// errors
var (
	ErrEmailTaken   = core.NErr("account with ths email already exist")
	ErrNameTaken    = core.NErr("name is already taken")
	ErrNotVerified  = core.NErr("account is not verified")
	ErrNotFound     = core.NErr("not found")
	ErrInvalidLogin = core.NErr("password or name is incorrect")
	ErrNotAuthor    = core.NErr("you cannot edit note you are not author of")
)

// DB is main database interface
type DB struct {
	Client *mongo.Client

	*mongo.Database

	Ctx context.Context

	Cancel context.CancelFunc

	Accounts, Notes, CounterA, CounterN *mongo.Collection

	vCodeFactory
}

// NDB sets up a database
func NDB(clientAddress, name string) (rdb *DB, err error) {
	if clientAddress == "default" {
		clientAddress = "mongodb://127.0.0.1:27017"
	}
	db := DB{vCodeFactory: *nVCodeFactory()}
	db.Client, err = mongo.NewClient(options.Client().ApplyURI(clientAddress))
	if err != nil {
		return
	}

	db.Ctx, db.Cancel = context.WithCancel(context.Background())
	err = db.Client.Connect(db.Ctx)
	if err != nil {
		return
	}

	db.Database = db.Client.Database(name)

	db.Accounts = db.Collection(Accounts)
	_, err = db.Accounts.Indexes().CreateMany(db.Ctx, MakeIndex(AccountIndex))
	if err != nil {
		panic(err)
	}

	db.Notes = db.Collection(Notes)
	_, err = db.Notes.Indexes().CreateMany(db.Ctx, MakeIndex(NoteIndex))
	if err != nil {
		panic(err)
	}

	db.CounterA = db.Collection(CounterA)
	db.CounterN = db.Collection(CounterN)

	rdb = &db

	return
}

// IDCounter stores incremented id
type IDCounter struct {
	ID    core.ID `bson:"_id"`
	Value core.ID
}

// NID creates new unique incremental id
func (d *DB) NID(counter *mongo.Collection) core.ID {
	var c IDCounter
	err := counter.FindOne(d.Ctx, All).Decode(&c)
	if err == mongo.ErrNoDocuments {
		counter.InsertOne(d.Ctx, IDCounter{})
		return 0
	} else if err != nil {
		panic(err)
	}

	_, err = counter.UpdateOne(d.Ctx, All, Inc(bson.M{"value": 1}))
	if err != nil {
		panic(err)
	}

	return c.Value + 1
}

// Drop drops the database, after this DB cannot be used
func (d *DB) Drop() {
	d.Database.Drop(d.Ctx)
	d.Cancel()
}

// AccountByID reads account from database, returns false if account wos not found
func (d *DB) AccountByID(id core.ID) (ac core.Account, err error) {
	err = d.Accounts.FindOne(d.Ctx, bson.M{"_id": id}).Decode(&ac)
	err = AssertNotFound(err)
	return
}

// AccountByEmail finds account based of a email, email of every account has to be unique
func (d *DB) AccountByEmail(email string) (ac core.Account, err error) {
	err = d.Accounts.FindOne(d.Ctx, bson.M{"email": email}).Decode(&ac)
	err = AssertNotFound(err)
	return
}

// AccountByName finds account based of a name, name of every account has to be unique
func (d *DB) AccountByName(name string) (ac core.Account, err error) {
	err = d.Accounts.FindOne(d.Ctx, bson.M{"name": name}).Decode(&ac)
	err = AssertNotFound(err)

	return
}

// AccountIdsForName collects all account ids witch name starts with given string
func (d *DB) AccountIdsForName(name string) (ids []core.RawID, err error) {
	c, err := d.Accounts.Find(d.Ctx, bson.D{StartsWith("name", name)})
	if err != nil {
		return
	}
	err = c.All(d.Ctx, &ids)
	return
}

// LoginAccount returns account with given password and name
func (d *DB) LoginAccount(name, password string) (ac core.Account, err error) {
	err = d.Accounts.FindOne(d.Ctx, bson.M{"name": name, "password": password}).Decode(&ac)
	err = AssertNotFound(err)
	if err != nil {
		err = ErrInvalidLogin
		return
	}

	if ac.Code != Verified {
		err = ErrNotVerified
	}
	return
}

//UpdateNoteList ...
func (d *DB) UpdateNoteList(id core.ID, list []core.ID) error {
	_, err := d.Accounts.UpdateOne(d.Ctx, ID(id), Set(bson.M{"Notes": list}))
	return err
}

// UpdateAccount replaces account with modified one
func (d *DB) UpdateAccount(ac *core.Account) error {
	_, err := d.Accounts.ReplaceOne(d.Ctx, ID(ac.ID), ac)
	return err
}

// AddAccount inserts account to database, also generates id, if name is already taken,
// account is not inserted and false is returned
func (d *DB) AddAccount(ac *core.Account) error {
	_, err := d.AccountByEmail(ac.Email)
	if err == nil {
		return ErrEmailTaken
	}

	_, err = d.AccountByName(ac.Name)
	if err == nil {
		return ErrNameTaken
	}

	ac.ID = d.NID(d.CounterA)
	ac.Code = d.vCodeFactory.value()

	_, err = d.Accounts.InsertOne(d.Ctx, ac)
	if err != nil {
		panic(err)
	}

	return nil
}

// ChangeAccountCode is used when user enters incorrect code to prevent brute force attacks
func (d *DB) ChangeAccountCode(id core.ID) string {
	code := d.vCodeFactory.value()
	_, err := d.Accounts.UpdateOne(d.Ctx, ID(id), Set(bson.M{"code": code}))
	if err != nil {
		panic(err)
	}

	return code
}

// MakeAccountVerified is used when user enters correct code to clarify that account is now verified
func (d *DB) MakeAccountVerified(id core.ID) {
	_, err := d.Accounts.UpdateOne(d.Ctx, ID(id), Set(bson.M{"code": Verified}))
	if err != nil {
		panic(err)
	}
}

// DraftBySID abstracts id parsing step
func (d *DB) DraftBySID(sid string) (dr core.Draft, err error) {
	id, err := core.ParseID(sid)
	if err != nil {
		err = core.ErrImpossible.Wrap(err)
		return
	}
	dr, err = d.DraftByID(id)
	return
}

// DraftByID ...
func (d *DB) DraftByID(id core.ID) (dr core.Draft, err error) {
	err = d.Notes.FindOne(d.Ctx, ID(id)).Decode(&dr)
	err = AssertNotFound(err)
	return
}

// NoteBySID abstracts id parsing step
func (d *DB) NoteBySID(sid string) (n core.Note, err error) {
	id, err := core.ParseID(sid)
	if err != nil {
		err = core.ErrImpossible.Wrap(err)
		return
	}
	n, err = d.NoteByID(id)
	return
}

// NoteByID ...
func (d *DB) NoteByID(id core.ID) (n core.Note, err error) {
	err = d.Notes.FindOne(d.Ctx, ID(id)).Decode(&n)
	err = AssertNotFound(err)
	return
}

// SetPublished ...
func (d *DB) SetPublished(id core.ID, value bool) error {
	_, err := d.Notes.UpdateOne(d.Ctx, ID(id), Set(bson.M{"published": value}))
	return err
}

// IsAuthor returns ErrNotAuthor if given note has different author
func (d *DB) IsAuthor(owner, note core.ID) error {
	_, err := d.Notes.Find(d.Ctx, bson.M{"_id": note, "owner": owner})
	err = AssertNotFound(err)
	if err != nil {
		err = ErrNotAuthor
	}
	return err
}

// AddNote inserts note to database, also generates id
func (d *DB) AddNote(nt *core.Note) {
	nt.ID = d.NID(d.CounterN)
	nt.BornDate = time.Now().UnixNano() / int64(time.Microsecond)
	_, err := d.Notes.InsertOne(d.Ctx, nt)
	if err != nil {
		panic(err)
	}
}

// UpdateNote overwrites note with its modified version, target is determinate by id
func (d *DB) UpdateNote(nt *core.Note) {
	_, err := d.Notes.ReplaceOne(d.Ctx, ID(nt.ID), nt)
	if err != nil {
		panic(err)
	}
}

// SearchNote returns fitting search results for given parameters
func (d *DB) SearchNote(values url.Values, published bool) []core.NotePreview {
	res, err := d.Notes.Find(d.Ctx, d.NoteFilter(values, published))
	if err != nil {
		panic(err)
	}

	notes := make([]core.NotePreview, 0, MaxCursorSize)
	for i := 0; res.TryNext(d.Ctx) && i <= MaxCursorSize; i++ {
		notes = notes[:i+1]
		res.Decode(&notes[i])
		if len(notes[i].Content) > MaxPreviewSize {
			notes[i].Content = notes[i].Content[:MaxPreviewSize]
		}
	}

	return notes
}

type vCodeFactory struct {
	r rand.Rand
	m sync.Mutex
}

func nVCodeFactory() *vCodeFactory {
	src := rand.NewSource(int64(time.Now().Nanosecond()))

	return &vCodeFactory{
		r: *rand.New(src),
	}
}

func (v *vCodeFactory) value() string {
	v.m.Lock()
	defer v.m.Unlock()

	return strconv.Itoa(v.r.Intn(999999))
}

// AssertNotFound makes sure error is equal to mongo.ErrNoDocuments and returns more user friendly ErrNotFound
func AssertNotFound(err error) error {
	if err != nil {
		if err != mongo.ErrNoDocuments {
			panic(err)
		}

		return ErrNotFound
	}

	return nil
}
