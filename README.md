# Fader

[![Status](https://img.shields.io/badge/status-dev-blue.svg)]()
[![License](http://img.shields.io/badge/license-mit-blue.svg)](https://raw.githubusercontent.com/fasder2/platform/master/LICENSE)
[![Build Status](https://travis-ci.org/fader2/platform.svg?branch=master)](https://travis-ci.org/fader2/platform)

Fader is a abstract tool for creation of web applications. Gives maximum flexibility arranging information on the site.

## Overview

## Features

## Quick start
```shell
# gen keys
openssl genrsa -out _key.pem 2048
openssl rsa -in _key.pem -pubout -out _key.pem.pub

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

## Support

team@fader.site

## Administrator's Guide

### Reload settings

To reset the settings of the running application

```shell
# ps -ax | grep -i platform
kill -SIGUSR1 PID
```

# Contributing

Fader is 100% free and open source. We encourage and support an active, healthy community that accepts contributions from the public – including you!

# Security

We take security very seriously at Fader; all our code is 100% open source and peer reviewed. Please read our security guide for an overview of security measures in Fader, or if you wish to report a security issue.