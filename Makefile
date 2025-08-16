TESTOUT=/tmp/test-out
BIN=cms/cms

all: build unittests filetest
	true

build:
	cd cms && go build .

unittests: build
	cd cms/core && go test

filetest: unittests build
	rm -rf ${TESTOUT}/*
	mkdir -p ${TESTOUT}
	${BIN} dump templates/business-card-01 --out=${TESTOUT}/business-card-01
	diff -r -q ${TESTOUT}/business-card-01 tests/business-card-01.golden
	echo "All tests passed successfully"

run: build
	${BIN} run templates/business-card-01
