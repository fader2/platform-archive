# Addons

To extend the functionality of the application.

## Quick start

```shell
# add new addons
$ ./cli.sh newaddon cache

$ tree -L 2  addons/cache/
addons/cache/
```

```
├── Makefile
├── addon.go
├── assets
│   ├── generate.go
│   └── templates
└── views
```