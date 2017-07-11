local boltdb = require("boltdb")
ctx:set("HelpLink", not (ctx:queryParam("dev") == ""))



-- Если нет cookie создать и сохранить их
-- В контексте хранить флаг определяющий были ли изменения в куках - сохранять ли при окончании процедуры
-- При запросе брать по ID значение пользователя
-- При завершении обновлять если надо

-- u = user.new()
-- u:SetMeta("a", "B")
-- print("user is exists", u:IsExists())
-- print("user meta", u:Meta("a"))
-- u:SetPassword("123")
-- print("user match pwd", u:MatchPassword("123"))
-- print("user match pwd", u:MatchPassword("1234"))
-- print("user match pwd", u:MatchPassword(""))
-- sessionToken = ctx():cookieValue("_token")

-- if (sessionToken == nil) then
--     token = GenToken("8b984f00-892f-4be2-97cc-80a60238a7fd")
--     print("new token", token)
--     ctx():setCookie("_token", token)
-- else 
--     store = boltdb.Bucket("users")
--     user = Auth(store, sessionToken)
--     print("session token", user:IsExists())
--     -- print("delete token")
--     -- ctx():DelCookie("_token")
-- end

token = accessToken.generate("8b984f00-892f-4be2-97cc-80a60238a7fd")
print(token)
id = accessToken.decode(token)
print(id)

store = boltdb.bucket("users")

egor, exists = user.findByLogin(store, "egor")

if not exists then
    print("not exists egor - create new user")
    egor = user.new("egor")
end

print("empty user - is guest?", egor:isGuest())
print("empty user - meta[a]", egor:meta("a"))
egor:meta("a", "b")
print("empty user - meta[a]", egor:meta("a"))
print("empty user - meta[a]", egor:meta("a"))
egor:pwd("1234")
print("empty user - match pwd", egor:matchPwd("a"))
print("empty user - match pwd", egor:matchPwd("1234"))
egor:save(store)

egor:meta("basket", join({"a", 1, "c"}, ","))

ctx:session(egor)

do return end
