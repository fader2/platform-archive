local foo = require("foo")
local boltdb = require("boltdb")

foo.Init() -- init external extensions

-- cfg():Dev(true)
cfg():AddRoute("GET", "/", "pages/index.jet", {"_all.lua"})

-- basic
cfg():Set(DOMAIN, "www.domain.com")
-- settings cookies
cfg():Set(DEF_COOKIE_EXPIRES, "8640h") -- 365 days

boltdb.Init()
boltdb.Opens({"sessions", "templates", "users"})