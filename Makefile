VERBOSE=-v

all: grammar test release

grammar:
	go generate ./...

release: spiff_linux_amd64.zip spiff_darwin_amd64.zip spiff_linux_ppc64le.zip

linux:
	GOOS=linux GOARCH=amd64 go build -o spiff++/spiff++ .

test:
	go test $(VERBOSE) ./...

spiff_linux_amd64.zip:
	GOOS=linux GOARCH=amd64 go build -o spiff++/spiff++ .
	rm -f spiff++/spiff_linux_amd64.zip
	(cd spiff++; zip spiff_linux_amd64.zip spiff++)
	rm spiff++/spiff++

spiff_darwin_amd64.zip:
	GOOS=darwin GOARCH=amd64 go build -o spiff++/spiff++ .
	rm -f spiff++/spiff_spiff_darwin_amd64.zip
	(cd spiff++; zip spiff_darwin_amd64.zip spiff++)
	rm spiff++/spiff++

spiff_linux_ppc64le.zip:
	GOOS=linux GOARCH=ppc64le go build -o spiff++/spiff++ .
	rm -f spiff++/spiff_linux_ppc64le.zip
	(cd spiff++; zip spiff_linux_ppc64le.zip spiff++)
	rm spiff++/spiff++

vendor:
	go mod vendor

clean:
	rm -rf ./spiff++

tidy:
	go mod tidy
	# restore patched version of candiedyaml/decode.go
	#git checkout -- vendor/github.com/cloudfoundry-incubator/candiedyaml/decode.go
	#git checkout -- vendor/github.com/cloudfoundry-incubator/candiedyaml/emitter.go
