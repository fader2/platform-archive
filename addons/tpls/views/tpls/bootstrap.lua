print("bootstrap tpls")
cfg():Set("tpls", "init OK")

cfg():AddRoute("GET", "/console/tpls/fragments/:fragmentID/edit", "tpls/fragment_edit.jet", "tpls/fragment_edit.lua")
cfg():AddRoute("POST", "/console/tpls/fragments/:fragmentID/edit", "", "tpls/fragment_edit.lua")