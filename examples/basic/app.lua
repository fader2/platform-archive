local foo = require("foo")
local boltdb = require("boltdb")
local tpls = require("tpls")

-- foo.init() -- init external extensions

print(cfg:get("foo"))
-- cfg():Dev(true)
cfg:addRoute("GET", "/", "pages/index.jet", {"_all.lua", "pages/index.lua"})
cfg:addRoute("GET", "/signin", "pages/login.jet", {"_all.lua"})

-- settings cookies
cfg:set(COOKIE_SECURE, false)

boltdb.init()
boltdb.opens({"sessions", "templates", "users"})

tpls.init()