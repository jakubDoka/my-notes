package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"myNotes/core"
	"myNotes/core/mongo"
	"net/http"
	"strings"

	"github.com/jakubDoka/gogen/str"
	"github.com/jakubDoka/sterr"
	"github.com/jakubDoka/urlp"
)

const success = "success"

// Error message constants for debuging
var (
	ErrAccount           = sterr.New("failed to create account")
	ErrSendEmail         = sterr.New("failed to send verification email (please report)")
	ErrInvalidLogin      = sterr.New("failed to login")
	ErrIncorrectCode     = sterr.New("code you entered is incorrect, we sent you email with new code for next try")
	ErrAlreadyVerified   = sterr.New("your account is already verified")
	ErrIllegalNoteAccess = sterr.New("you are not an author of this notes so you cannot edit them")
	ErrInvalidBoolean    = sterr.New("invalid boolean")
	ErrInvalidUserCookie = sterr.New("user cookie is invalid")
	ErrMissingUserCookie = sterr.New("missing user cookie")
	ErrNotPublished      = sterr.New("this note is not published yet")
)

// WS like a website, struct is main interface to frontend, it opens a server and handels requests
type WS struct {
	db            *mongo.DB
	fs            http.Handler
	targetAddress string
	bot           EmailSender
	ps            urlp.Parser
}

// NWS creates new WS that can then be runned by ws.Run()
func NWS(domain, pageDir string, port int16, db *mongo.DB, bot EmailSender) (nws *WS) {
	return &WS{
		db:            db,
		fs:            http.FileServer(http.Dir(pageDir)),
		targetAddress: fmt.Sprintf("%s:%d", domain, port),
		bot:           bot,
		ps:            urlp.New(urlp.LowerCase),
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
	http.HandleFunc("/publicaccount", w.PublicAccount)
	http.HandleFunc("/config", w.Config)
	http.HandleFunc("/configure", w.Configure)
	// note
	http.HandleFunc("/search", w.Search)
	http.HandleFunc("/save", w.SaveNote)
	http.HandleFunc("/publicnote", w.PublicNote)
	http.HandleFunc("/privatenote", w.PrivateNote)
	http.HandleFunc("/usernotes", w.UserNotes)
	http.HandleFunc("/setpublished", w.SetPublished)
	// general
	http.HandleFunc("/like", w.Like)
}

// RegisterAccount handels registering account and responds whether registration wos successful
func (w *WS) RegisterAccount(wr http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	encoder, failed := w.Setup(wr, r, &req)
	if failed {
		return
	}

	ac := core.Account{
		Name:     req.Name,
		Password: req.Password,
		Email:    req.Email,
	}

	err := func() (err error) {
		err = ValidEmail(ac.Email)
		if err != nil {
			return
		}

		err = w.db.CanCreateAccount(&ac)
		if err != nil {
			return ErrAccount.Wrap(err)
		}

		ac.Code = w.db.Code()

		err = w.SendVerifycationEmail(&ac)
		if err != nil {
			return
		}

		err = w.db.Account(&ac)
		if err != nil {
			return core.EI(err)
		}

		return
	}()

	encoder.Encode(NResponce(err))
}

// VerifyAccount ...
func (w *WS) VerifyAccount(wr http.ResponseWriter, r *http.Request) {
	var req VerifyRequest
	encoder, failed := w.Setup(wr, r, &req)
	if failed {
		return
	}

	err := func() (err error) {
		ac, err := w.db.LoginAccount(req.Name, req.Password)
		if err != nil && !errors.Is(err, mongo.ErrNotVerified) {
			return ErrInvalidLogin.Wrap(err)
		}

		err = nil

		if ac.Code == mongo.Verified {
			return ErrAlreadyVerified
		}

		if ac.Code != req.Code {
			ac.Code, err = w.db.ChangeAccountCode(ac.ID)
			if err != nil {
				return
			}

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

// PublicAccount retrieves account from db, but only by id and password is censored
func (w *WS) PublicAccount(wr http.ResponseWriter, r *http.Request) {
	var req IDRequest
	encoder, failed := w.Setup(wr, r, &req)
	if failed {
		return
	}

	ac, err := w.db.AccountByID(req.ID)

	ac.Censure()

	encoder.Encode(AccountResponce{
		Resp:    NResponce(err),
		Account: ac,
	})
}

// Account retrieves account from db
func (w *WS) Account(wr http.ResponseWriter, r *http.Request) {
	var req Request
	encoder, failed := w.Setup(wr, r, &req)
	if failed {
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
	var req core.SearchRequest
	encoder, failed := w.Setup(wr, r, &req)
	if failed {
		return
	}

	if req.Author == "!!me" {
		if ac, err := w.GetAccountFromCookie(wr, r); err == nil {
			req.Author = mongo.ExactLabel + ac.Name
		}
	}

	res, err := w.db.SearchNote(req, true)

	if len(res) == 0 {
		err = mongo.ErrNotFound
	}

	encoder.Encode(SearchResponce{
		Resp:    NResponce(err),
		Results: res,
	})
}

// Login creates cookies to remember user
func (w *WS) Login(wr http.ResponseWriter, r *http.Request) {
	var req LoginReqest
	encoder, failed := w.Setup(wr, r, &req)
	if failed {
		return
	}

	ac, err := w.db.LoginAccount(req.Name, req.Password)
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
	var req = OptIDRequest{ID: core.None}
	encoder, failed := w.Setup(wr, r, &req)
	if failed {
		return
	}

	var ac core.Account
	var err error
	if req.ID != core.None {
		ac, err = w.db.AccountByID(req.ID)
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
	var req ConfigureRequest
	encoder, failed := w.Setup(wr, r, &req)
	if failed {
		return
	}

	err := func() (err error) {
		ac, err := w.GetAccountFromCookie(wr, r)
		if err != nil {
			return
		}

		if req.Name != ac.Name {
			ac.Name = req.Name
			_, err = w.db.AccountByName(req.Name)
			if err == nil {
				return mongo.ErrNameTaken
			}
			err = nil
		}

		ac.Cfg.Colors = strings.Split(req.Colors, " ")
		for i, c := range ac.Cfg.Colors {
			ac.Cfg.Colors[i] = "#" + c
		}

		err = w.db.Replace(w.db.Accounts, &ac)
		if err != nil {
			return
		}

		// name changes so cookie has to be restored
		cookie := ac.Cookie()
		http.SetCookie(wr, &cookie)

		return
	}()

	encoder.Encode(NResponce(err))
}

// Like changes like state of something
func (w *WS) Like(wr http.ResponseWriter, r *http.Request) {
	var req LikeRequest
	encoder, failed := w.Setup(wr, r, &req)
	if failed {
		return
	}

	var state bool
	var amount int
	err := func() (err error) {
		ac, err := w.GetAccountFromCookie(wr, r)
		if err != nil {
			return
		}

		tp, err := core.ParseTargetType(req.Target)
		if err != nil {
			return
		}

		state, amount, err = w.db.Like(req.ID, ac.ID, w.db.Coll(tp), req.Change)
		return
	}()

	encoder.Encode(LikeResponce{
		Resp:  NResponce(err),
		State: state,
		Count: amount,
	})
}

// Comment adds comment to note or reply to commant
func (w *WS) Comment(wr http.ResponseWriter, r *http.Request) {
	var req CommentRequest
	encoder, failed := w.Setup(wr, r, &req)
	if failed {
		return
	}

	err := func() (err error) {
		ac, err := w.GetAccountFromCookie(wr, r)
		if err != nil {
			return
		}

		act, err := w.db.TakeAction(ac.ID)
		if err != nil {
			return
		}

		bytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return core.EI(err)
		}

		cm := core.Comment{
			Author:  ac.ID,
			Note:    core.None,
			Content: string(bytes),
		}

		tp, err := core.ParseTargetType(req.Target)
		if err != nil {
			return
		}

		cm.Target.Type = tp
		switch tp {
		case core.CommentT:
			nt, err := w.db.NoteByID(req.ID)
			if err != nil {
				return err
			}
			cm.Target.ID = nt.ID
		case core.NoteT:
			ocm, err := w.db.CommentByID(req.ID)
			if err != nil {
				return err
			}
			cm.Target.ID = ocm.ID
		}

		err = w.db.Comment(&cm)
		if err != nil {
			return
		}

		return act()
	}()

	encoder.Encode(NResponce(err))
}

// SaveNote creates new note if id == "new" it creates new note otherwise, it just updates data
func (w *WS) SaveNote(wr http.ResponseWriter, r *http.Request) {
	var req = SaveRequest{ID: core.None}
	encoder, failed := w.Setup(wr, r, &req)
	if failed {
		return
	}

	var (
		ac   core.Account
		note = core.Note{
			Year:    req.Year,
			Month:   req.Month,
			Name:    req.Name,
			School:  mongo.School(req.School),
			Subject: req.Subject,
			Theme:   req.Theme,
		}
	)

	err := func() (err error) {
		ac, err = w.GetAccountFromCookie(wr, r)
		if err != nil {
			return
		}

		var act = func() error { return nil }
		if req.ID == core.None {
			act, err = w.db.TakeAction(ac.ID)
			if err != nil {
				return
			}

			note.Author = ac.ID
		} else {
			note, err = w.db.NoteByID(req.ID)
			if err != nil {
				return err
			}
		}

		bytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return core.EI(err)
		}

		note.Content = string(bytes)

		if req.ID == core.None {
			err = w.db.Note(&note)
		} else {
			err = w.db.UpdateNote(&note)
		}

		if err != nil {
			return
		}

		return act()
	}()

	encoder.Encode(SaveResponce{
		Resp: NResponce(err),
		ID:   note.ID,
	})
}

// SetPublished alters publicity of note
func (w *WS) SetPublished(wr http.ResponseWriter, r *http.Request) {
	var req PublishRequest
	encoder, failed := w.Setup(wr, r, &req)
	if failed {
		return
	}

	err := func() (err error) {
		ac, err := w.GetAccountFromCookie(wr, r)
		if err != nil {
			return
		}

		err = w.db.IsAuthor(ac.ID, req.ID)
		if err != nil {
			return
		}

		err = w.db.SetPublished(req.ID, req.Publish)
		if err != nil {
			panic(err)
		}

		return
	}()

	encoder.Encode(NResponce(err))
}

// PrivateNote retrieves any note but only if user is an owner of a note
func (w *WS) PrivateNote(wr http.ResponseWriter, r *http.Request) {
	w.Note(wr, r, true)
}

// PublicNote returns any public note but not private
func (w *WS) PublicNote(wr http.ResponseWriter, r *http.Request) {
	w.Note(wr, r, false)
}

// Note retrieves note by id
func (w *WS) Note(wr http.ResponseWriter, r *http.Request, private bool) {
	var req IDRequest
	encoder, failed := w.Setup(wr, r, &req)
	if failed {
		return
	}

	var nt core.Note

	err := func() (err error) {

		nt, err = w.db.NoteByID(req.ID)
		if err != nil {
			return
		}

		if private {
			ac, err := w.GetAccountFromCookie(wr, r)
			if err != nil {
				return err
			}

			if ac.ID != nt.Author {
				return ErrIllegalNoteAccess
			}
		} else if !nt.Published {
			return ErrNotPublished
		}

		return
	}()

	encoder.Encode(NoteResponce{
		Resp: NResponce(err),
		Note: nt,
	})
}

// UserNotes retrieves all notes user has as drafts
func (w *WS) UserNotes(wr http.ResponseWriter, r *http.Request) {
	var req IDRequest
	encoder, failed := w.Setup(wr, r, &req)
	if failed {
		return
	}

	var nts []core.Draft
	err := w.db.UserNotes(req.ID, &nts)

	encoder.Encode(DraftResponce{
		Resp:   NResponce(err),
		Drafts: nts,
	})
}

// SendVerifycationEmail creates the message and sends it to targeted account
func (w *WS) SendVerifycationEmail(account *core.Account) error {
	message := FormatVerificationEmail(account.Code, account.Name)
	return w.bot.Send(message, account.Email)
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

// Setup handles invalid request, it sends error responce to sender and returns nil if all
// asserted arguments weren't inputted
func (w *WS) Setup(wr http.ResponseWriter, r *http.Request, request interface{}) (*json.Encoder, bool) {
	err := w.ps.Parse(r.URL.Query(), request)

	if err != nil {
		http.Error(wr, err.Error(), http.StatusBadRequest)
		return nil, true
	}

	wr.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(wr), false
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
