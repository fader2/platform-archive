package consts

const (
	DOMAIN             = "__domain"             // string
	DEF_COOKIE_EXPIRES = "__def_cookie_expires" // 24h
	DEF_COOKIE_SECURE  = "__dev_cookie_secure"  // bool

	// место хранения виджетов
	// в случае boltdb это бакет
	// в случае других хранилищ это может быть таблицой, tuple и тп
	TPL_WIDGETS_BUCKET_NAME = "__tpl_widgets_bucket"
)
