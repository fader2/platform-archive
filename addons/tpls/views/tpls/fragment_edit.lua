local boltdb = require("boltdb")
local store = boltdb.bucket(TPL_FRAGMENTS_BUCKET_NAME)

if ctx:isPost() then
    local newData = ctx:formValue("data")
    local fragmentID = ctx:formValue("fragmentID")
    store:set(fragmentID, newData)
    ctx:redirect("/console/tpls/fragments/"..fragmentID.."/edit")
    do return end
end

local data = store:get(ctx:get("fragmentID"))
ctx:set("Data", data) 
ctx:set("NotExists", data == nil)
