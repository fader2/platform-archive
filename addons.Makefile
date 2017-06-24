build_addons:
	make -C addons/foo build
	make -C addons/boltdb build 
.PHONY: build_addons