SHELL = /bin/sh

VERSION=0.9.5
BUILD=`git rev-parse HEAD`

LDFLAGS=-ldflags "-w -s \
				-X github.com/Betterment/testtrack-cli/cmds.version=${VERSION} \
				-X github.com/Betterment/testtrack-cli/cmds.build=${BUILD}"

PACKAGES=$$(find . -maxdepth 1 -type d ! -path '.' ! -path './.*' ! -path './vendor' ! -path './dist' ! -path './script' ! -path './doc')
LINTEXCLUDES="123nothingyet123"

all: test

install:
	@go install ${LDFLAGS} github.com/Betterment/testtrack-cli/testtrack

dist:
	@mkdir dist &&\
		GOOS=linux GOARCH=amd64 go build -o "dist/testtrack.linux" ${LDFLAGS} github.com/Betterment/testtrack-cli/testtrack &&\
		GOOS=darwin GOARCH=amd64 go build -o "dist/testtrack.darwin" ${LDFLAGS} github.com/Betterment/testtrack-cli/testtrack

release: distclean dist
	@hub release create\
		-a dist/testtrack.linux\
		-a dist/testtrack.darwin\
		-m "TestTrack CLI ${VERSION}"\
		-t "${BUILD}"\
		v${VERSION}

test:
	@go install ./vendor/golang.org/x/tools/cmd/goimports
	@go install ./vendor/github.com/golang/lint/golint
	@GOIMPORTS_RESULT=$$(goimports -l ${PACKAGES} | grep -v ${LINTEXCLUDES});\
		if [ $$(echo "$$GOIMPORTS_RESULT\c" | head -c1 | wc -c) -ne 0 ];\
			then\
				echo "Style violations found. Run the following command to fix:";\
				echo;\
				echo "goimports -w" $$GOIMPORTS_RESULT;\
				echo;\
				exit 1;\
			fi
	@go vet ${PACKAGES}
	@GOLINT_RESULT=$$(golint ${PACKAGES} | grep -v ${LINTEXCLUDES});\
		if [ $$(echo "$$GOLINT_RESULT\c" | head -c1 | wc -c) -ne 0 ];\
			then\
				echo $$GOLINT_RESULT;\
				exit 1;\
			fi
	@go test ${PACKAGES}

cover:
	@echo "What package do you want a coverage report for? \c"
	@read PACKAGE &&\
		go test --coverprofile=cover.out ./$$PACKAGE
	@go tool cover -html=cover.out

coverall:
	@echo "${PACKAGES}" | xargs -I {} -n 1 sh -c 'go test --coverprofile=cover.out {} | grep -v '?' && go tool cover -html=cover.out'

# Clean up all compiled sources
clean:
	@go clean ./...

# Clean everything except pgp keys
distclean: clean
	@rm -rf dist

.PHONY: all build install check clean distclean test
