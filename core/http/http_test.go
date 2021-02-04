package http

import (
	"encoding/json"
	"myNotes/core"
	"myNotes/core/mongo"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestWSRegisterAccount(t *testing.T) {
	db, ws := SetupTest()
	defer db.Cancel()

	testCases := []struct {
		desc   string
		args   url.Values
		result Responce
	}{
		{
			desc: "successfull",
			args: url.Values{
				"n": {"name"},
				"p": {"password"},
				"e": {"jakub.doka2@gmail.com"},
			},
			result: Responce{success},
		},
		{
			desc: "invalid email",
			args: url.Values{
				"n": {"name"},
				"p": {"password"},
				"e": {"emailthatdoesnotexist24567@gmail.com"},
			},
			result: Responce{ErrInvalidEmail.Error()},
		},
		{
			desc: "name taken",
			args: url.Values{
				"n": {"name"},
				"p": {"password"},
				"e": {"mlokogrgel@gmail.com"},
			},
			result: Responce{ErrAccount.Wrap(mongo.ErrNameTaken).Error()},
		},
		{
			desc: "email taken",
			args: url.Values{
				"n": {"name"},
				"p": {"password"},
				"e": {"jakub.doka2@gmail.com"},
			},
			result: Responce{ErrAccount.Wrap(mongo.ErrEmailTaken).Error()},
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, DoTest("register", tC.args, tC.result, ws.RegisterAccount))
	}
}

func TestVerify(t *testing.T) {
	db, ws := SetupTest()
	defer db.Cancel()

	ac := core.Account{
		Name:     "name",
		Password: "password",
		Email:    "mlokogrgel@gmail.com",
	}

	db.Account(&ac)
	db.Account(&core.Account{
		Name:     "name1",
		Password: "password",
		Email:    "jakub.doka2@gmail.com",
	})

	testCases := []struct {
		desc   string
		args   url.Values
		result Responce
	}{
		{
			desc: "login fail",
			args: url.Values{
				"n": {"name"},
				"p": {""},
				"c": {ac.Code},
			},
			result: Responce{ErrInvalidLogin.Wrap(mongo.ErrInvalidLogin).Error()},
		},
		{
			desc: "successfull",
			args: url.Values{
				"n": {"name"},
				"p": {"password"},
				"c": {ac.Code},
			},
			result: Responce{success},
		},
		{
			desc: "already",
			args: url.Values{
				"n": {"name"},
				"p": {"password"},
				"c": {ac.Code},
			},
			result: Responce{ErrAlreadyVerified.Error()},
		},
		{
			desc: "incorrect code",
			args: url.Values{
				"n": {"name1"},
				"p": {"password"},
				"c": {""},
			},
			result: Responce{ErrIncorrectCode.Error()},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, DoTest("verify", tC.args, tC.result, ws.VerifyAccount))
	}
}

func TestLogin(t *testing.T) {
	db, ws := SetupTest()
	defer db.Cancel()

	_ = MakeVerifiedAccount(db)

	db.Account(&core.Account{
		Name:     "name",
		Password: "password",
		Email:    "jakub.doka2@gmail.com",
	})
	testCases := []struct {
		desc   string
		args   url.Values
		result Responce
	}{
		{
			desc: "incorrect name",
			args: url.Values{
				"n": {""},
				"p": {"password"},
			},
			result: Responce{ErrInvalidLogin.Wrap(mongo.ErrInvalidLogin).Error()},
		},
		{
			desc: "incorrect password",
			args: url.Values{
				"n": {"name"},
				"p": {""},
			},
			result: Responce{ErrInvalidLogin.Wrap(mongo.ErrInvalidLogin).Error()},
		},
		{
			desc: "not verified",
			args: url.Values{
				"n": {"name"},
				"p": {"password"},
			},
			result: Responce{ErrInvalidLogin.Wrap(mongo.ErrNotVerified).Error()},
		},
		{
			desc: "successful",
			args: url.Values{
				"n": {"name1"},
				"p": {"password"},
			},
			result: Responce{success},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, DoTest("login", tC.args, tC.result, ws.Login))
	}
}

func TestAccount(t *testing.T) {
	db, ws := SetupTest()
	defer db.Cancel()

	ac := MakeVerifiedAccount(db)

	testCases := []struct {
		desc   string
		result AccountResponce
	}{
		{
			desc: "success",
			result: AccountResponce{
				Resp:    Responce{success},
				Account: ac,
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, DoTest("account", url.Values{}, tC.result, ws.Account, ac.Cookie()))
	}
}

func TestConfig(t *testing.T) {
	db, ws := SetupTest()
	defer db.Cancel()

	ac := MakeVerifiedAccount(db)

	testCases := []struct {
		desc   string
		result ConfigResponce
		cookie http.Cookie
	}{
		{
			desc: "success",
			result: ConfigResponce{
				Resp: Responce{success},
				Cfg:  ac.Cfg,
			},
			cookie: ac.Cookie(),
		},
		{
			desc: "invalid data",
			result: ConfigResponce{
				Resp: Responce{ErrInvalidUserCookie.Error()},
				Cfg:  core.Config{},
			},
			cookie: http.Cookie{Name: "user"},
		},
		{
			desc: "invalid data",
			result: ConfigResponce{
				Resp: Responce{ErrMissingUserCookie.Error()},
				Cfg:  core.Config{},
			},
			cookie: http.Cookie{},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, DoTest("config", url.Values{}, tC.result, ws.Config, tC.cookie))
	}
}

func TestConfigure(t *testing.T) {
	db, ws := SetupTest()
	defer db.Cancel()

	ac := MakeVerifiedAccount(db)

	db.Account(&core.Account{
		Name: "name2",
	})

	testCases := []struct {
		desc   string
		args   url.Values
		result Responce
		cookie http.Cookie
	}{
		{
			desc: "invalid name",
			args: url.Values{
				"n": {"name2"},
				"c": {""},
			},
			result: Responce{mongo.ErrNameTaken.Error()},
			cookie: ac.Cookie(),
		},

		{
			desc: "success",
			args: url.Values{
				"n": {"name3"},
				"c": {""},
			},
			result: Responce{success},
			cookie: ac.Cookie(),
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, DoTest("configure", tC.args, tC.result, ws.Configure, tC.cookie))
	}
}

func DoTest(callback string, args url.Values, result interface{}, call func(w http.ResponseWriter, r *http.Request), cookies ...http.Cookie) func(t *testing.T) {
	return func(t *testing.T) {
		rc := httptest.NewRecorder()
		for _, c := range cookies {
			http.SetCookie(rc, &c)
		}

		req, err := http.NewRequest("GET", "/"+callback+"?"+args.Encode(), nil)
		if err != nil {
			panic(err)
		}

		req.Header["Cookie"] = rc.HeaderMap["Set-Cookie"]

		handler := http.HandlerFunc(call)

		handler.ServeHTTP(rc, req)
		if rc.Code != http.StatusOK {
			t.Error(rc.Code, http.StatusOK, "bad status")
		}

		bts, err := json.Marshal(result)
		if err != nil {
			panic(err)
		}

		if rc.Body.String() != string(bts)+"\n" {
			t.Error(rc.Body.String(), string(bts))
		}
	}

}

func MakeVerifiedAccount(db *mongo.DB) core.Account {
	ac := core.Account{
		Name:     "name1",
		Password: "password",
		Email:    "mlokogrgel@gmail.com",
	}
	db.Account(&ac)
	db.MakeAccountVerified(ac.ID)
	ac.Code = mongo.Verified

	return ac
}

func SetupTest() (*mongo.DB, *WS) {
	db, err := mongo.NDB("default", "test")
	if err != nil {
		panic(err)
	}

	db.Drop()

	db, err = mongo.NDB("default", "test")
	if err != nil {
		panic(err)
	}

	bot := NEmailSender("mlokogrgel@gmail.com", "mlokMLOK1234", 587)

	ws := NWS("127.0.0.1", "./web", 3000, db, *bot)

	return db, ws
}
