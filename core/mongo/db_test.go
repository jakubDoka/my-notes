package mongo

import (
	"myNotes/core"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
)

func TestNID(t *testing.T) {
	db := Setup()

	for i := core.ID(0); i < 4; i++ {
		id := db.NID(db.CounterN)
		if id != i {
			t.Errorf("%d != %d", id, i)
		}
	}
}

func TestInsert(t *testing.T) {

	db := Setup()

	ac := core.Account{
		Name:     "guhuhu",
		Password: "buhuhu",
	}

	err := db.AddAccount(&ac)
	if err != nil {
		panic(err)
	}

	ac2, _ := db.AccountByID(ac.ID)

	if ac2.Name != ac.Name || ac2.Password != ac.Password {
		t.Errorf("%v != %v", ac, ac2)
	}

	db.Drop()
}

func TestDifTypes(t *testing.T) {
	db := Setup()

	db.AddNote(&core.Note{
		Name:    "hello",
		Content: "hello there am sprighstea jsd sa",
	})

	var ac struct {
		Name, Content string
	}
	err := db.Notes.FindOne(db.Ctx, bson.M{}).Decode(&ac)
	if err != nil {
		t.Error(ac, err)
	}

}

func Setup() *DB {
	db, err := NDB("default", "test")
	if err != nil {
		panic(err)
	}
	db.Drop()
	db, err = NDB("default", "test")
	if err != nil {
		panic(err)
	}

	return db
}
