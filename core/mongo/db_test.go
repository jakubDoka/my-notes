package mongo

import (
	"myNotes/core"
	"strconv"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
)

func TestNID(t *testing.T) {
	db := Setup()

	for i := core.ID(0); i < 4; i++ {
		id, err := db.NID()
		if id != i || err != nil {
			t.Errorf("%d != %d, %v", id, i, err)
		}
	}

	for i := core.ID(0); i < 4; i++ {
		err := db.DID(i)
		if err != nil {
			t.Error(err)
		}
	}

	for i := core.ID(4); i > 0; i-- {
		id, err := db.NID()
		if id != i-1 || err != nil {
			t.Errorf("%d != %d, %v", id, i-1, err)
		}
	}
}

func TestInsert(t *testing.T) {

	db := Setup()

	ac := core.Account{
		Name:     "guhuhu",
		Password: "buhuhu",
	}

	err := db.Account(&ac)
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

	db.Note(&core.Note{
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

func TestSearch(t *testing.T) {
	db := Setup()
	acs := []core.Account{
		{Name: "hh"},
		{Name: "hhb"},
		{Name: "hk"},
		{Name: "ah"},
	}

	for i := range acs {
		acs[i].Email = strconv.Itoa(i)
		db.Account(&acs[i])
	}

	nts := []core.Note{
		{Name: "aa", Author: 1, Year: 2, Month: 3, Theme: "a", Subject: "f", School: 0},
		{Name: "aab", Author: 1, Year: 1, Month: 5, Theme: "a", Subject: "g", School: 0},
		{Name: "bb", Author: 2, Year: 5, Month: 6, Theme: "fa", Subject: "g", School: 0},
		{Name: "bc", Author: 0, Year: 2, Month: 6, Theme: "ca", Subject: "fa", School: 0},
	}

	for i := range nts {
		db.Note(&nts[i])
	}

	testCases := []struct {
		desc    string
		query   core.SearchRequest
		results []core.NotePreview
	}{
		{
			desc:  "no filter",
			query: core.SearchRequest{},
			results: []core.NotePreview{
				{Name: "aa", Author: 1, ID: 0},
				{Name: "aab", Author: 1, ID: 1},
				{Name: "bb", Author: 2, ID: 2},
				{Name: "bc", Author: 0, ID: 3},
			},
		},

		{
			desc:  "exact author",
			query: core.SearchRequest{Author: "!hh"},
			results: []core.NotePreview{
				{Name: "bc", Author: 0, ID: 3},
			},
		},

		{
			desc:  "author",
			query: core.SearchRequest{Author: "hh"},
			results: []core.NotePreview{
				{Name: "aa", Author: 1, ID: 0},
				{Name: "aab", Author: 1, ID: 1},
				{Name: "bc", Author: 0, ID: 3},
			},
		},

		{
			desc: "regular",
			query: core.SearchRequest{
				Name:   "aa",
				Theme:  "a",
				Author: "hh",
			},
			results: []core.NotePreview{
				{Name: "aa", Author: 1, ID: 0},
				{Name: "aab", Author: 1, ID: 1},
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			res, err := db.SearchNote(tC.query, false)
			if err != nil {
				t.Error(err)
				return
			}
			if len(res) != len(tC.results) {
				t.Error(res, "!=", tC.results)
				return
			}

			mp := map[string]bool{}
			for _, r := range tC.results {
				mp[r.String()] = true
			}

			for i := range res {
				if !mp[res[i].String()] {
					t.Error(res, "!=", tC.results)
				}
			}
		})
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
