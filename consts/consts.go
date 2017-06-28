package consts

import "github.com/fader2/platform/config"

const (
	DOMAIN          = "__domain"          // string
	COOKIE_EXPIRES  = "__cookie_expires"  // string time.DurationParse
	COOKIE_SECURE   = "__cookie_secure"   // bool
	COOKIE_SID_NAME = "__cookie_sid_name" // string

	// место хранения фрагментов
	// в случае boltdb это бакет
	// в случае других хранилищ это может быть таблицой, tuple и тп
	TPL_FRAGMENTS_BUCKET_NAME = "__tpl_fragments_bucket" // string

	JWT_EXPIRATION_DELTA = "__jwt_expiration_delta" // string time.DurationParse

	HTTP_GET  = "GET"
	HTTP_POST = "POST"
)

var (
	DEF_DOMAIN          = ".localhost"
	DEF_COOKIE_EXPIRES  = "8640h" // 365 days
	DEF_COOKIE_SECURE   = true
	DEF_COOKIE_SID_NAME = "TOK"

	DEF_TPL_FRAGMENTS_BUCKET_NAME = "tpl_fragments_bucket"
	DEF_JWT_EXPIRATION_DELTA      = "8640h" // 365 days
)

func SetupDefValues(c *config.Config) {
	c.Vars.Set(DOMAIN, DEF_DOMAIN)
	c.Vars.Set(COOKIE_EXPIRES, DEF_COOKIE_EXPIRES)
	c.Vars.Set(COOKIE_SECURE, DEF_COOKIE_SECURE)
	c.Vars.Set(TPL_FRAGMENTS_BUCKET_NAME, DEF_TPL_FRAGMENTS_BUCKET_NAME)
	c.Vars.Set(JWT_EXPIRATION_DELTA, DEF_JWT_EXPIRATION_DELTA)
	c.Vars.Set(COOKIE_SID_NAME, DEF_COOKIE_SID_NAME)
}
