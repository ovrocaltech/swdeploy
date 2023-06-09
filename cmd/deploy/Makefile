# Makefile for go. Passes in Version and git hash.

BINARY=swdeploy
REMOTE=nginxservice.lwa.pvt:/var/www/html/deployments/

VERSION=`git describe --dirty`
BUILD=`git rev-parse HEAD`
TAG=`git tag --points-at HEAD`

LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.Build=${BUILD}"

build:
	go build ${LDFLAGS} -o ${BINARY}

build-arm:
	env GOOS=linux GOARCH=arm go build ${LDFLAGS} -o ${BINARY}-arm

install:
	go install ${LDFLAGS}
	tar cf ../${BINARY}_${TAG}.tar.gz ../deploy
	rsync -av -e ssh ../${BINARY}_${TAG}.tar.gz ${REMOTE}

clean:
	if [ -f ${BINARY} ] ; then rm ${BINARY} ; fi
	if [ -f ${BINARY}-arm ] ; then rm ${BINARY}-arm ; fi

.PHONY: clean install

