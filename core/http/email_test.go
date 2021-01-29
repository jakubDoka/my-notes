package http

import (
	"encoding/json"
	"errors"
	"strconv"
	"testing"
)

func TestValidEmail(t *testing.T) {
	tests := []struct {
		email  string
		result error
	}{
		{"aro@gma.cmo", ErrInvalidEmail},
		{"mlokogrgel@gmail.com", nil},
	}

	for i, te := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			res := ValidEmail(te.email)
			if !errors.Is(res, te.result) {
				t.Error(res, te.result)
			}
		})
	}
}

func TestEmailSender(t *testing.T) {
	s := NEmailSender(BotAccount.Email, BotAccount.Password, 587)

	err := s.Send([]byte("Hello there."), "jakub.doka2@gmail.com")
	if err != nil {
		t.Error(err)
	}
}

// testing whether marshaler accepts anonymous struct
func TestResponce(t *testing.T) {
	m, err := json.Marshal(struct {
		Hello     int
		Something string
	}{10, "Hello"})
	if err != nil {
		t.Error(err)
	}
	t.Error(string(m))
}
