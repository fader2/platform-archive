local boltdb = require("boltdb")
boltdb.Opens({TPL_FRAGMENTS_BUCKET_NAME})

print("bootstrap boltdb")
cfg():Set("boltdb", true)