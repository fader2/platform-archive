local boltdb = require("boltdb")
local store = boltdb.Bucket(TPL_FRAGMENTS_BUCKET_NAME)

if ctx():IsPost() then
    local newData = ctx():FormValue("data")
    local fragmentID = ctx():FormValue("fragmentID")
    store:Set(fragmentID, newData)
    ctx():Redirect("/console/tpls/fragments/"..fragmentID.."/edit")
    do return end
end

local data = store:Get(ctx():Get("fragmentID"))
ctx():Set("Data", data) 
ctx():Set("NotExists", data == nil)
