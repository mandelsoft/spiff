GODEPS := $(shell godep path)
GOPATH := $(GODEPS):$(GOPATH)

grammar:
	go get github.com/pointlander/peg
	peg dynaml/dynaml.peg

release: spiff_linux_amd64.zip spiff_darwin_amd64.zip	

linux:
	GOPATH=$(GOPATH) GOOS=linux GOARCH=amd64 go build -o spiff++/spiff++ .

spiff_linux_amd64.zip:
	GOPATH=$(GOPATH) GOOS=linux GOARCH=amd64 go build -o spiff++/spiff++ .
	rm -f spiff++/spiff_linux_amd64.zip
	(cd spiff++; zip spiff_linux_amd64.zip spiff++)
	rm spiff++/spiff++

spiff_darwin_amd64.zip:
	GOPATH=$(GOPATH) GOOS=darwin GOARCH=amd64 go build -o spiff++/spiff++ .
	rm -f spiff++/spiff_spiff_darwin_amd64.zip
	(cd spiff++; zip spiff_darwin_amd64.zip spiff++)
	rm spiff++/spiff++
