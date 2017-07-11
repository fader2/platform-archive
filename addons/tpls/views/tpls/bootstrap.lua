print("bootstrap tpls")
cfg:set("tpls", "init OK")

cfg:addRoute("GET", "/console/tpls/fragments/:fragmentID/edit", "tpls/fragment_edit.jet", "tpls/fragment_edit.lua")
cfg:addRoute("POST", "/console/tpls/fragments/:fragmentID/edit", "", "tpls/fragment_edit.lua")