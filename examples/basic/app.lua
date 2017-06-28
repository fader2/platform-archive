local foo = require("foo")
local boltdb = require("boltdb")
local tpls = require("tpls")

foo.Init() -- init external extensions

-- cfg():Dev(true)
cfg():AddRoute("GET", "/", "pages/index.jet", {"_all.lua"})

-- settings cookies
cfg():Set(COOKIE_SECURE, false)

boltdb.Init()
boltdb.Opens({"sessions", "templates", "users"})

tpls.Init()