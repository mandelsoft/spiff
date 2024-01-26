VERBOSE=-v

all: vendor grammar test release

grammar:
	go generate ./...

release: spiff_linux_amd64.zip spiff_darwin_amd64.zip spiff_linux_ppc64le.zip spiff_linux_arm64.zip spiff_windows_amd64.zip spiff_windows_386.zip

linux:
	GOOS=linux GOARCH=amd64 go build -o spiff++/spiff++ .

test:
	go test $(VERBOSE) --count=1 ./...

spiff_linux_amd64.zip:
	GOOS=linux GOARCH=amd64 go build -o spiff++/spiff++ .
	rm -f spiff++/spiff_linux_amd64.zip
	(cd spiff++; zip spiff_linux_amd64.zip spiff++)
	rm spiff++/spiff++

spiff_linux_arm64.zip:
	GOOS=linux GOARCH=arm64 go build -o spiff++/spiff++ .
	rm -f spiff++/spiff_linux_arm64.zip
	(cd spiff++; zip spiff_linux_arm64.zip spiff++)
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

spiff_windows_amd64.zip:
	GOOS=windows GOARCH=amd64 go build -o spiff++/spiff++.exe .
	rm -f spiff++/spiff_windows_amd64.zip
	(cd spiff++; zip spiff_windows_amd64.zip spiff++.exe)
	rm spiff++/spiff++.exe

spiff_windows_386.zip:
	GOOS=windows GOARCH=386 go build -o spiff++/spiff++_386.exe .
	rm -f spiff++/spiff_windows_386.zip
	(cd spiff++; zip spiff_windows_386.zip spiff++_386.exe)
	rm spiff++/spiff++_386.exe

.PHONY: vendor
vendor:
	go mod vendor

clean:
	rm -rf ./spiff++

tidy:
	go mod tidy
	# restore patched version of candiedyaml/decode.go
	#git checkout -- vendor/github.com/cloudfoundry-incubator/candiedyaml/decode.go
	#git checkout -- vendor/github.com/cloudfoundry-incubator/candiedyaml/emitter.go
