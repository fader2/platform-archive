build_addons:
	make -C addons/foo build
	make -C addons/boltdb build 
	make -C addons/tpls build 
.PHONY: build_addons