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

Add builder to `addons.Makefile`.

``` diff
build_addons:
	make -C addons/foo build
	make -C addons/boltdb build
+   # new addon
+   make -C addons/cache build // new addon
.PHONY: build_addons
```