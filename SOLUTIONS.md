## Quick solutions

* User authorization
* User registration

### User registration

```lua
-- from form
local boldb = require("boltdb") -- or other storage
store = boltdb.Bucket("users")

```

### User authorization

``` lua
-- authorization form
local boltdb = require("boltdb") -- or other storage
store = boltdb.Bucket("users") -- заранее подготовленная таблица\бакет\спейс в зависимости от хранилища

-- find by login
userID = boltdb:GetRefID("user_name_from_form")
user = FindUserByID(store, userID)
if user:MatchPassword(pwd) then
    print("auth")
end

-- find by ID
user = FindUserByID(store, userID)

-- find by Token
-- token = GenToken("8b984f00-892f-4be2-97cc-80a60238a7fd")
user = Auth(store, token)
```