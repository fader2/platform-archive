# the lua environment

NOTE: `any` type it is `number|bool|string|userdata|table`

## Context

`ctx:<function_name>`

Methods

* get(key string) any
* set(key string, value any)
* setSessionUser
* sessionUser
* queryParam(name string) string
* formValue(name string) string
* path() string
* cookieValue(name string) string|nil
* setCookie(name string, value string)
* delCookie(name string)
* status(httpStatus int)
* noContent(httpStatus int)
* redirect(url string)
* json(httpStatus int, value any)
* blob(httpStatus int, contentType string, data userdata)
* isPost() bool
* isGet() bool
* isDelete() bool

## Global methods

* GenToken(uuid string) string - генерирует токен привязанный к UUID
* 