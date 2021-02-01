package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"gogen/str"
	"io/ioutil"
	"log"
	"myNotes/core"
	"myNotes/core/mongo"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const success = "success"

// Error message constants for debuging
var (
	ErrAccount           = core.NErr("failed to create account")
	ErrSendEmail         = core.NErr("failed to send verification email (please report)")
	ErrInvalidLogin      = core.NErr("failed to login")
	ErrIncorrectCode     = core.NErr("code you entered is incorrect, we sent you email with new code for next try")
	ErrAlreadyVerified   = core.NErr("your account is already verified")
	ErrIllegalNoteAccess = core.NErr("you are not an author of this notes so you cannot edit them")
	ErrInvalidBoolean    = core.NErr("invalid boolean")
	ErrInvalidUserCookie = core.NErr("user cookie is invalid")
	ErrMissingUserCookie = core.NErr("missing user cookie")
)

// WS like a website, struct is main interface to frontend, it opens a server and handels requests
type WS struct {
	db            *mongo.DB
	fs            http.Handler
	targetAddress string
	bot           EmailSender
}

// NWS creates new WS that can then be runned by ws.Run()
func NWS(domain, pageDir string, port int16, db *mongo.DB, bot EmailSender) (nws *WS) {
	return &WS{
		db:            db,
		fs:            http.FileServer(http.Dir(pageDir)),
		targetAddress: fmt.Sprintf("%s:%d", domain, port),
		bot:           bot,
	}
}

// RegisterHandlers ...
func (w *WS) RegisterHandlers() {
	http.Handle("/", w.fs)
	// account
	http.HandleFunc("/register", w.RegisterAccount)
	http.HandleFunc("/verify", w.VerifyAccount)
	http.HandleFunc("/login", w.Login)
	http.HandleFunc("/account", w.Account)
	http.HandleFunc("/config", w.Config)
	http.HandleFunc("/configure", w.Configure)
	// note
	http.HandleFunc("/search", w.Search)
	http.HandleFunc("/save", w.SaveNote)
	http.HandleFunc("/note", w.Note)
	http.HandleFunc("/draft", w.Draft)
	http.HandleFunc("/setpublished", w.SetPublished)
}

// RegisterAccount handels registering account and responds whether registration wos successful
func (w *WS) RegisterAccount(wr http.ResponseWriter, r *http.Request) {
	args, encoder := Setup(wr, r, SetupAssert{"n": 1, "p": 1, "e": 1})
	if args == nil {
		return
	}

	account := core.Account{
		Name:     args["n"][0],
		Password: args["p"][0],
		Email:    args["e"][0],
	}

	err := func() (err error) {
		err = ValidEmail(account.Email)
		if err != nil {
			return
		}

		err = w.db.AddAccount(&account)
		if err != nil {
			return ErrAccount.Wrap(err)
		}

		err = w.SendVerifycationEmail(&account)
		if err != nil {
			return
		}

		return
	}()

	encoder.Encode(NResponce(err))
}

// VerifyAccount ...
func (w *WS) VerifyAccount(wr http.ResponseWriter, r *http.Request) {
	args, encoder := Setup(wr, r, SetupAssert{"n": 1, "p": 1, "c": 1})
	if args == nil {
		return
	}

	var (
		name     = args["n"][0]
		password = args["p"][0]
		code     = args["c"][0]
	)

	err := func() (err error) {
		ac, err := w.db.LoginAccount(name, password)
		if err != nil && !errors.Is(err, mongo.ErrNotVerified) {
			return ErrInvalidLogin.Wrap(err)
		}

		err = nil

		if ac.Code == mongo.Verified {
			return ErrAlreadyVerified
		}

		if ac.Code != code {
			ac.Code = w.db.ChangeAccountCode(ac.ID)
			err = w.SendVerifycationEmail(&ac)
			if err != nil {
				return
			}
			return ErrIncorrectCode
		}

		w.db.MakeAccountVerified(ac.ID)

		return
	}()

	encoder.Encode(NResponce(err))
}

// Account retrieves account from db
func (w *WS) Account(wr http.ResponseWriter, r *http.Request) {
	args, encoder := Setup(wr, r, SetupAssert{})
	if args == nil {
		return
	}

	ac, err := w.GetAccountFromCookie(wr, r)
	encoder.Encode(AccountResponce{
		Resp:    NResponce(err),
		Account: ac,
	})
}

// Search searches notes based of url params
func (w *WS) Search(wr http.ResponseWriter, r *http.Request) {
	args, encoder := Setup(wr, r, SetupAssert{
		"name":    1,
		"month":   1,
		"school":  1,
		"subject": 1,
		"theme":   1,
		"year":    1,
		"author":  1,
	})

	if args == nil {
		return
	}

	if args["author"][0] == "#me" {
		if ac, err := w.GetAccountFromCookie(wr, r); err == nil {
			args["author"][0] = "#" + ac.Name
		}
	}

	res := w.db.SearchNote(args, true)

	status := success

	if len(res) == 0 {
		status = "nothing found"
	}

	encoder.Encode(SearchResponce{
		Resp:    Responce{status},
		Results: res,
	})
}

// Login creates cookies to remember user
func (w *WS) Login(wr http.ResponseWriter, r *http.Request) {
	args, encoder := Setup(wr, r, SetupAssert{"n": 1, "p": 1})
	if args == nil {
		return
	}

	var (
		name     = args["n"][0]
		password = args["p"][0]
	)

	ac, err := w.db.LoginAccount(name, password)
	if err != nil {
		err = ErrInvalidLogin.Wrap(err)
	} else {
		cookie := ac.Cookie()
		http.SetCookie(wr, &cookie)
	}

	encoder.Encode(NResponce(err))
}

// Config returns user config to frontend based of cookie
func (w *WS) Config(wr http.ResponseWriter, r *http.Request) {
	args, encoder := Setup(wr, r, SetupAssert{})
	if args == nil {
		return
	}

	var ac core.Account
	var err error
	if val, ok := args["id"]; ok && len(val) != 0 {
		id, err := core.ParseID(val[0])
		if err != nil {
			ac, err = w.db.AccountByID(id)
		}
	} else {
		ac, err = w.GetAccountFromCookie(wr, r)
	}

	encoder.Encode(ConfigResponce{
		Resp: NResponce(err),
		Cfg:  ac.Cfg,
	})
}

// Configure changes user configuration
func (w *WS) Configure(wr http.ResponseWriter, r *http.Request) {
	args, encoder := Setup(wr, r, SetupAssert{"n": 1, "c": 1})
	if args == nil {
		return
	}

	err := func() (err error) {
		ac, err := w.GetAccountFromCookie(wr, r)
		if err != nil {
			return
		}

		name := args["n"][0]

		if name != ac.Name {
			ac.Name = name
			_, err = w.db.AccountByName(name)
			if err == nil {
				return mongo.ErrNameTaken
			}
			err = nil
		}

		ac.Cfg.Colors = strings.Split(args["c"][0], " ")
		for i, c := range ac.Cfg.Colors {
			ac.Cfg.Colors[i] = "#" + c
		}

		w.db.UpdateAccount(&ac)
		cookie := ac.Cookie()
		http.SetCookie(wr, &cookie)

		return
	}()

	encoder.Encode(NResponce(err))
}

// SaveNote creates new note if id == "new" it creates new note otherwise, it just updates data
func (w *WS) SaveNote(wr http.ResponseWriter, r *http.Request) {
	args, encoder := Setup(wr, r, SetupAssert{
		"name":    1,
		"month":   1,
		"school":  1,
		"subject": 1,
		"theme":   1,
		"year":    1,
		"id":      1,
	})
	if args == nil {
		return
	}

	var (
		ac   core.Account
		note core.Note
	)

	err := func() (err error) {
		ac, err = w.GetAccountFromCookie(wr, r)
		if err != nil {
			return
		}

		sid := args["id"][0]
		if sid == "new" {
			note.Author = ac.ID
			w.db.AddNote(&note)
			ac.Notes = append(ac.Notes, note.ID)
		} else {
			note, err = w.db.NoteBySID(sid)
			if err != nil {
				return err
			}
		}

		note.Year, err = strconv.Atoi(args["year"][0])
		if err != nil {
			return core.ErrImpossible.Wrap(err)
		}

		note.Month, err = strconv.Atoi(args["month"][0])
		if err != nil {
			return core.ErrImpossible.Wrap(err)
		}

		bytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return
		}
		note.Content = string(bytes)

		note.Name = args["name"][0]
		note.Theme = args["theme"][0]
		note.Subject = args["subject"][0]
		note.School = mongo.School(args["school"][0])

		w.db.UpdateNote(&note)
		w.db.UpdateNoteList(ac.ID, ac.Notes)

		return
	}()

	encoder.Encode(SaveResponce{
		Resp: NResponce(err),
		ID:   note.ID,
	})
}

// SetPublished alters publicity of note
func (w *WS) SetPublished(wr http.ResponseWriter, r *http.Request) {
	args, encoder := Setup(wr, r, SetupAssert{"id": 1, "b": 1})
	if args == nil {
		return
	}

	err := func() (err error) {
		ac, err := w.GetAccountFromCookie(wr, r)
		if err != nil {
			return
		}

		id, err := core.ParseID(args["id"][0])
		if err != nil {
			return core.ErrImpossible.Wrap(err)
		}

		err = w.db.IsAuthor(ac.ID, id)
		if err != nil {
			return
		}

		val, err := ParseBool(args["b"][0])
		if err != nil {
			return core.ErrImpossible.Wrap(err)
		}

		err = w.db.SetPublished(id, val)
		if err != nil {
			panic(err)
		}

		return
	}()

	encoder.Encode(NResponce(err))
}

// Note retrieves note by id
func (w *WS) Note(wr http.ResponseWriter, r *http.Request) {
	args, encoder := Setup(wr, r, SetupAssert{"id": 1})
	if args == nil {
		return
	}

	var nt core.Note

	err := func() (err error) {
		ac, err := w.GetAccountFromCookie(wr, r)
		if err != nil {
			return
		}

		nt, err = w.db.NoteBySID(args["id"][0])
		if err != nil {
			return
		}

		if ac.ID != nt.Author {
			return ErrIllegalNoteAccess
		}

		return
	}()

	encoder.Encode(NoteResponce{
		Resp: NResponce(err),
		Note: nt,
	})
}

// Draft retrieves draft data
func (w *WS) Draft(wr http.ResponseWriter, r *http.Request) {
	args, encoder := Setup(wr, r, SetupAssert{"id": 1})
	if args == nil {
		return
	}

	d, err := w.db.DraftBySID(args["id"][0])

	encoder.Encode(DraftResponce{
		Resp:  NResponce(err),
		Draft: d,
	})
}

// SendVerifycationEmail ...
func (w *WS) SendVerifycationEmail(account *core.Account) error {
	message := FormatVerificationEmail(account.Code, account.Name)
	err := w.bot.Send(message, account.Email)
	if err != nil {
		return ErrSendEmail.Wrap(err)
	}

	return nil
}

// Run launches the WS, server will be running until this method exits ends
func (w *WS) Run() {
	fmt.Println("server listening on", w.targetAddress)
	err := http.ListenAndServe(w.targetAddress, nil)
	if err != nil {
		log.Fatal(err)
	}
}

// GetAccountFromCookie extracts account from request cookie, cookie can be missing of value can be invalid do
// appropriate error is returned
func (w *WS) GetAccountFromCookie(wr http.ResponseWriter, r *http.Request) (ac core.Account, err error) {
	n, p, err := GetUserDataFromCookie(wr, r)
	if err != nil {
		return
	}

	ac, err = w.db.LoginAccount(n, p)
	if err != nil {
		if !errors.Is(err, mongo.ErrNotVerified) {
			err = ErrInvalidUserCookie
		}
		return
	}
	return
}

// SearchResponce ...
type SearchResponce struct {
	Resp    Responce
	Results []core.NotePreview
}

// AccountResponce ...
type AccountResponce struct {
	Resp    Responce
	Account core.Account
}

// ConfigResponce ...
type ConfigResponce struct {
	Resp Responce
	Cfg  core.Config
}

// DraftResponce ...
type DraftResponce struct {
	Resp  Responce
	Draft core.Draft
}

// SaveResponce ...
type SaveResponce struct {
	Resp Responce
	ID   core.ID
}

// NoteResponce ...
type NoteResponce struct {
	Resp Responce
	Note core.Note
}

// Responce is responce sent by RegisterAccount callback
type Responce struct {
	Status string
}

// NResponce creates new responce from error, if error is nil it substitutes success message
func NResponce(err error) Responce {
	status := success
	if err != nil {
		status = err.Error()
	}
	return Responce{status}
}

// SetupAssert stores parameters that input has to fit, determinate witch kes has to be present
// and how match elements they have to contain
type SetupAssert map[string]int

// Setup handles invalid request, it sends error responce to sender and returns nil if all
// asserted arguments weren't inputted
func Setup(w http.ResponseWriter, r *http.Request, assert SetupAssert) (url.Values, *json.Encoder) {
	val := r.URL.Query()
	var incorrect string
	for a, l := range assert {
		if val, ok := val[a]; ok {
			if len(val) != l {
				incorrect += fmt.Sprintf(" invalid amount of args: expected:%d got:%d", l, len(val))
			}
		} else {
			incorrect += " missing:" + a
		}
	}

	if len(incorrect) != 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid Url params:" + incorrect))
		return nil, nil
	}

	w.Header().Set("Content-Type", "application/json")
	return val, json.NewEncoder(w)
}

// GetUserDataFromCookie ...
func GetUserDataFromCookie(w http.ResponseWriter, r *http.Request) (name, password string, err error) {
	cookie, err := r.Cookie("user")
	if err != nil {
		err = ErrMissingUserCookie
		return
	}

	name, password = str.SplitToTwo(cookie.Value, ' ')
	return
}

// InternalErr reports internal server error
func InternalErr(w http.ResponseWriter, err string) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(err))
}

// ParseBool ...
func ParseBool(raw string) (bool, error) {
	switch raw {
	case "true":
		return true, nil
	case "false":
		return false, nil
	default:
		return false, ErrInvalidBoolean
	}
}
