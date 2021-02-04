package mongo

import (
	"context"
	"math/rand"
	"myNotes/core"
	"strconv"
	"sync"
	"time"

	"github.com/jakubDoka/sterr"
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

	CommentIndex = []string{
		"target.id",
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
	ErrEmailTaken   = sterr.New("account with ths email already exist")
	ErrNameTaken    = sterr.New("name is already taken")
	ErrNotVerified  = sterr.New("account is not verified")
	ErrNotFound     = sterr.New("not found")
	ErrInvalidLogin = sterr.New("password or name is incorrect")
	ErrNotAuthor    = sterr.New("you cannot edit note you are not author of")
	ErrLimmitRate   = sterr.New("you have to wait %s to take another action")
)

// DB is main database interface
type DB struct {
	Client *mongo.Client

	*mongo.Database

	Ctx context.Context

	Cancel context.CancelFunc

	Accounts, Notes, Comments, CounterA, CounterN, CounterC *mongo.Collection

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
func (d *DB) NID(counter *mongo.Collection) (core.ID, error) {
	var c IDCounter
	err := counter.FindOne(d.Ctx, All).Decode(&c)
	if err == mongo.ErrNoDocuments {
		counter.InsertOne(d.Ctx, IDCounter{})
		return 0, nil
	} else if err != nil {
		return 0, core.EI(err)
	}

	_, err = counter.UpdateOne(d.Ctx, All, Inc(bson.M{"value": 1}))
	if err != nil {
		return 0, core.EI(err)
	}

	return c.Value + 1, nil
}

// Coll returns collections based of target type
func (d *DB) Coll(tp core.TargetType) *mongo.Collection {
	switch tp {
	case core.CommentT:
		return d.Comments
	case core.NoteT:
		return d.Notes
	}

	panic("invalid TargetType")
}

// Replace replaces a document in given collection
func (d *DB) Replace(collection *mongo.Collection, doc core.IDer) error {
	_, err := collection.ReplaceOne(d.Ctx, ID(doc.AID()), doc)
	return core.EI(err)
}

// CheckLike returns whether document is liked by user
func (d *DB) CheckLike(id, user core.ID, collection *mongo.Collection) (liked bool, amount int, err error) {
	return d.Like(id, user, collection, false)
}

// ChangeLike likes or dislikes the document based of its current state
func (d *DB) ChangeLike(id, user core.ID, collection *mongo.Collection) (liked bool, amount int, err error) {
	return d.Like(id, user, collection, true)
}

// Like can change or return whether user has liked the document and optionally return id
func (d *DB) Like(id, user core.ID, collection *mongo.Collection, change bool) (liked bool, amount int, err error) {
	var likes core.Likes
	err = core.EI(collection.FindOne(d.Ctx, ID(id)).Decode(&likes))
	if err != nil {
		return
	}

	amount = len(likes.Likes)

	i, liked := likes.Likes.BiSearch(user, core.BiSearch)
	if change {
		if liked {
			_, err = collection.UpdateOne(d.Ctx, ID(id), Insert("likes", i, user))
			amount++
		} else {
			_, err = collection.UpdateOne(d.Ctx, ID(id), Pop("likes", i, 1))
			amount--
		}

		liked = !liked
	}

	return
}

// Drop drops the database, after this DB cannot be used
func (d *DB) Drop() {
	d.Database.Drop(d.Ctx)
	d.Cancel()
}

// AccountByID reads account from database, returns false if account wos not found
func (d *DB) AccountByID(id core.ID) (ac core.Account, err error) {
	err = d.Accounts.FindOne(d.Ctx, bson.M{"_id": id}).Decode(&ac)
	err = core.EI(err)
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
		return nil, core.EI(err)
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
	return core.EI(err)
}

// Account inserts account to database, also generates id, if name is already taken,
// account is not inserted and false is returned
func (d *DB) Account(ac *core.Account) error {
	_, err := d.AccountByEmail(ac.Email)
	if err == nil {
		return ErrEmailTaken
	}

	_, err = d.AccountByName(ac.Name)
	if err == nil {
		return ErrNameTaken
	}

	ac.ID, err = d.NID(d.CounterA)
	if err != nil {
		return nil
	}

	ac.Code = d.vCodeFactory.value()

	_, err = d.Accounts.InsertOne(d.Ctx, ac)
	if err != nil {
		return core.EI(err)
	}

	return nil
}

// ChangeAccountCode is used when user enters incorrect code to prevent brute force attacks
func (d *DB) ChangeAccountCode(id core.ID) (string, error) {
	code := d.vCodeFactory.value()
	_, err := d.Accounts.UpdateOne(d.Ctx, ID(id), Set(bson.M{"code": code}))
	return "", core.EI(err)
}

// MakeAccountVerified is used when user enters correct code to clarify that account is now verified
func (d *DB) MakeAccountVerified(id core.ID) error {
	_, err := d.Accounts.UpdateOne(d.Ctx, ID(id), Set(bson.M{"code": Verified}))
	return core.EI(err)
}

// TakeAction sets last action to current time
func (d *DB) TakeAction(id core.ID) (func() error, error) {
	ac, err := d.AccountByID(id)
	if err != nil {
		return nil, core.EI(err)
	}

	time := int64(core.ActionSpacing) - core.Time() - ac.LastAction

	if time > 0 {
		return nil, ErrLimmitRate.Args(core.FormatTime(time))
	}

	return func() error {
		_, err = d.Accounts.UpdateOne(d.Ctx, ID(id), Set(bson.M{"lastAction": core.Time()}))
		return core.EI(err)
	}, nil
}

// DraftByID ...
func (d *DB) DraftByID(id core.ID) (dr core.Draft, err error) {
	err = d.Notes.FindOne(d.Ctx, ID(id)).Decode(&dr)
	err = core.EI(err)
	return
}

// NoteByID ...
func (d *DB) NoteByID(id core.ID) (n core.Note, err error) {
	err = d.Notes.FindOne(d.Ctx, ID(id)).Decode(&n)
	err = core.EI(err)
	return
}

// SetPublished ...
func (d *DB) SetPublished(id core.ID, value bool) error {
	_, err := d.Notes.UpdateOne(d.Ctx, ID(id), Set(bson.M{"published": value}))
	return core.EI(err)
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

// Note inserts note to database, also generates id
func (d *DB) Note(nt *core.Note) (err error) {
	nt.ID, err = d.NID(d.CounterN)
	if err != nil {
		return
	}
	nt.BornDate = core.Time()
	_, err = d.Notes.InsertOne(d.Ctx, nt)
	return core.EI(err)
}

// UpdateNote overwrites note with its modified version, target is determinate by id
func (d *DB) UpdateNote(nt *core.Note) error {
	_, err := d.Notes.ReplaceOne(d.Ctx, ID(nt.ID), nt)
	return core.EI(err)
}

// SearchNote returns fitting search results for given parameters
func (d *DB) SearchNote(values core.SearchRequest, published bool) ([]core.NotePreview, error) {
	res, err := d.Notes.Find(d.Ctx, d.NoteFilter(values, published))
	if err != nil {
		return nil, core.EI(err)
	}

	notes := make([]core.NotePreview, 0, MaxCursorSize)
	for i := 0; res.TryNext(d.Ctx) && i <= MaxCursorSize; i++ {
		notes = notes[:i+1]
		res.Decode(&notes[i])
		if len(notes[i].Content) > MaxPreviewSize {
			notes[i].Content = notes[i].Content[:MaxPreviewSize]
		}
	}

	return notes, nil
}

// CommentByID ...
func (d *DB) CommentByID(id core.ID) (n core.Comment, err error) {
	err = d.Comments.FindOne(d.Ctx, ID(id)).Decode(&n)
	err = core.EI(err)
	return
}

// Comment adds new comment to db
func (d *DB) Comment(cm *core.Comment) (err error) {
	cm.ID, err = d.NID(d.CounterC)
	if err != nil {
		return core.EI(err)
	}
	cm.BornDate = core.Time()
	_, err = d.Comments.InsertOne(d.Ctx, cm)
	return core.EI(err)
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
			return core.EI(err)
		}

		return ErrNotFound
	}

	return nil
}
