
# Fader

Quick start
```shell
# create workspace
mkdir -p _workspace

cat >> _workspace/fader.lua << EOF
cfg():AddRoute("GET", "/hello/:name", "index.jet", "index.lua")
cfg():Dev(true)
EOF

cat >> _workspace/layout.jet << EOF
<!DOCTYPE html>
<meta charset="UTF-8">
<title>{{ isset(title) ? title : "n/a" }}</title>
<body>
{{block body()}}{{end}}
</body>
EOF

cat >> _workspace/index.jet << EOF
{{extends "layout.jet"}}
{{block body()}}
Name: {{name}}
{{end}}
EOF

cat >> _workspace/index.lua << EOF
ctx():Setx("title", "Title page")
ctx():Status(200)
EOF

make run

curl -s http://localhost:8383/hello/Fader2
<!DOCTYPE html>
<meta charset="UTF-8">
<title>Title page</title>
<body>

Name: Fader2

</body>
```

## Administrator's Guide

### Reload settings

To reset the settings of the running application

```shell
# ps -ax | grep -i platform
kill -SIGUSR1 PID
```