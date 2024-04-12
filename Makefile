.PHONY: clean build build-all pretty package

clean:
	@rm -rf dist

build:
	goreleaser build --snapshot --clean --single-target

build-all:
	goreleaser build --clean

pretty:
	@goimports -w cmd pkg sdk internal

package: clean pretty build-all
	goreleaser release --snapshot --clean
