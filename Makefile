build:
	go build

install: build
	mv tfgen /usr/local/bin
