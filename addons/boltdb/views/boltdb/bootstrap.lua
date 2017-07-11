local boltdb = require("boltdb")
boltdb.opens({TPL_FRAGMENTS_BUCKET_NAME})

print("bootstrap boltdb")
cfg:set("boltdb", true)