package objects

import "golang.org/x/crypto/bcrypt"

func (u *User) IsGuest() bool {
	return u.IsAccess("guest")
}

func (u *User) Type() string {
	return u.Meta.Meta[META_USER_TYPE]
}

func (u *User) IsAccess(in string) bool {
	for _, access := range u.Info.Pasport.Access {
		if access == in {
			return true
		}
	}
	return false
}

func (u *User) SetPWD(pwd string) {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	u.Info.Pasport.Pwd = string(hashedPassword)
}

func (u User) MatchPWD(pwd string) bool {
	return bcrypt.CompareHashAndPassword([]byte(u.Info.Pasport.Pwd), []byte(pwd)) == nil
}
