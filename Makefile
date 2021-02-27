VERBOSE=-v

all: grammar test release

grammar:
	go get github.com/pointlander/peg
	(cd $(GOPATH)/src/github.com/pointlander/peg; git checkout 1d0268dfff9bca9748dc9105a214ace2f5c594a8; go install .)
	peg dynaml/dynaml.peg

release: spiff_linux_amd64.zip spiff_darwin_amd64.zip spiff_linux_ppc64le.zip

linux: ensure
	GOOS=linux GOARCH=amd64 go build -o spiff++/spiff++ .

test: ensure
	go test $(VERBOSE) ./...

spiff_linux_amd64.zip: ensure
	GOOS=linux GOARCH=amd64 go build -o spiff++/spiff++ .
	rm -f spiff++/spiff_linux_amd64.zip
	(cd spiff++; zip spiff_linux_amd64.zip spiff++)
	rm spiff++/spiff++

spiff_darwin_amd64.zip: ensure
	GOOS=darwin GOARCH=amd64 go build -o spiff++/spiff++ .
	rm -f spiff++/spiff_spiff_darwin_amd64.zip
	(cd spiff++; zip spiff_darwin_amd64.zip spiff++)
	rm spiff++/spiff++

spiff_linux_ppc64le.zip: ensure
	GOOS=linux GOARCH=ppc64le go build -o spiff++/spiff++ .
	rm -f spiff++/spiff_linux_ppc64le.zip
	(cd spiff++; zip spiff_linux_ppc64le.zip spiff++)
	rm spiff++/spiff++

ensure:
	dep ensure
	# restore patched version of candiedyaml/decode.go
	#git checkout -- vendor/github.com/cloudfoundry-incubator/candiedyaml/decode.go
	#git checkout -- vendor/github.com/cloudfoundry-incubator/candiedyaml/emitter.go
