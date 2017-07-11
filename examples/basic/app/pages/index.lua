local sessionUser = ctx:session()
print(sessionUser==nil)
if (sessionUser == nil) then
    print("not exists user")
    ctx:noContent(403)
    do return end
end

print("len basket", #split(ctx:session():meta("basket"), ","))