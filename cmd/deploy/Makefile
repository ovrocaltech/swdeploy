# Makefile for go. Passes in Version and git hash.

BINARY=swdeploy

VERSION=`git describe --dirty`
BUILD=`git rev-parse HEAD`

LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.Build=${BUILD}"

build:
	go build ${LDFLAGS} -o ${BINARY}

build-arm:
	env GOOS=linux GOARCH=arm go build ${LDFLAGS} -o ${BINARY}-arm

install:
	go install ${LDFLAGS}

clean:
	if [ -f ${BINARY} ] ; then rm ${BINARY} ; fi
	if [ -f ${BINARY}-arm ] ; then rm ${BINARY}-arm ; fi

.PHONY: clean install

