build:
	go build -o bin/tfgen

install: build
	mv bin/tfgen /usr/local/bin
