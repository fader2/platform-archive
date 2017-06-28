package objects

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
