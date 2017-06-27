package objects

import (
	"testing"

	uuid "github.com/satori/go.uuid"
)

func TestUser(t *testing.T) {
	s := newTestStore()

	u := EmptyUser(Client)
	u.ID = uuid.NewV4()
	u.Info.Pasport.Email = "client@test.com"
	u.Info.Profile.FirstName = "fname"

	id, err := SetUser(s, u)
	if err != nil {
		t.Error("set user", err)
	}
	if id == uuid.Nil {
		t.Error("empty user ID")
	}

	//

	got, err := GetUser(s, id)
	if err != nil {
		t.Error("find user", err)
	}

	if !Client.Equal(UserType(got.Meta.Get(META_USER_TYPE))) {
		t.Error("not expected user type", got.Meta)
	}

	if got.Info.Pasport.Email != u.Info.Pasport.Email {
		t.Error("not expected email")
	}

	if got.Info.Profile.FirstName != u.Info.Profile.FirstName {
		t.Error("not expected first name")
	}
}