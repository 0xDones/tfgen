build:
	go build -o bin/tfgen

test:
	go test -v ./tfgen

install: build
	mv bin/tfgen /usr/local/bin
