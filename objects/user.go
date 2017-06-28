package objects

import (
	"bytes"
	io "io"

	"github.com/fader2/platform/consts"
	uuid "github.com/satori/go.uuid"
)

type UserType string

func (t UserType) String() string {
	return string(t)
}

func (t UserType) Equal(in UserType) bool {
	return t.String() == in.String()
}

const (
	META_USER_TYPE = "User-Type"
)

var (
	UnknownUserType UserType = ""
	Application     UserType = "app"
	Client          UserType = "client"
)

func EmptyUser(_type UserType) (u *User) {
	return &User{
		Meta: Meta{Meta: map[string]string{
			META_USER_TYPE: _type.String(),
		}},
		Info: UserInfo{
			Pasport: &UserPasport{},
			Profile: &UserProfile{},
		},
	}
}

type User struct {
	ID   uuid.UUID
	Meta Meta
	Info UserInfo
}

func GetUser(s Storer, id uuid.UUID) (*User, error) {
	o, err := s.EncodedObject(UserObject, id)
	if err != nil {
		return nil, err
	}

	return DecodeUser(o)
}

func SetUser(s Storer, b *User) (uuid.UUID, error) {
	obj := s.NewEncodedObject(b.ID)
	if err := b.Encode(obj); err != nil {
		return b.ID, err
	}
	return s.SetEncodedObject(obj)
}

func DecodeUser(o EncodedObject) (*User, error) {
	obj := EmptyUser(UnknownUserType)

	return obj, obj.Decode(o)
}

func (u *User) Decode(o EncodedObject) error {
	if o.Type() != UserObject {
		return consts.ErrNotSupported
	}

	u.ID = o.ID()
	u.Meta = o.Meta()

	buf := bytes.NewBuffer(make([]byte, 0, o.Size()))
	r, err := o.Reader()
	if err != nil {
		return err
	}
	defer r.Close()
	_, err = io.Copy(buf, r)
	if err != nil {
		return err
	}

	return u.Info.Unmarshal(buf.Bytes())
}

func (u *User) Encode(o EncodedObject) error {
	o.SetType(UserObject)
	o.SetMeta(u.Meta)
	w, err := o.Writer()
	if err != nil {
		return err
	}
	defer w.Close()
	dat, err := u.Info.Marshal()
	if err != nil {
		return err
	}
	_, err = io.Copy(w, bytes.NewReader(dat))
	return err
}
