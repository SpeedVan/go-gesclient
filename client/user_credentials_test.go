package client_test

import (
	"testing"

	"github.com/SpeedVan/go-gesclient/client"
)

func TestNewUserCredentials(t *testing.T) {
	var uc *client.UserCredentials
	uc = client.NewUserCredentials("user", "pswd")
	if uc == nil {
		t.Fail()
	}
}

func TestUserCredentials_Username(t *testing.T) {
	uc := client.NewUserCredentials("user", "pswd")
	if uc.Username() != "user" {
		t.Fail()
	}
}

func TestUserCredentials_Password(t *testing.T) {
	uc := client.NewUserCredentials("user", "pswd")
	if uc.Password() != "pswd" {
		t.Fail()
	}
}
