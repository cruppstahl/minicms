TESTOUT=/tmp/test-out
BIN=serve/serve

all: build test
	true

build:
	cd serve && go build .

test: build
	rm -rf ${TESTOUT}/*
	mkdir -p ${TESTOUT}
	${BIN} dump templates/business-card-01 --out=${TESTOUT}/business-card-01
	diff -r -q ${TESTOUT}/business-card-01 tests/business-card-01.golden

run: build
	${BIN} run templates/business-card-01
