grammar:
	go get github.com/pointlander/peg
	(cd $(GOPATH)/src/github.com/pointlander/peg; git checkout 1d0268dfff9bca9748dc9105a214ace2f5c594a8; go install .)
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
