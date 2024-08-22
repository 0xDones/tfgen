build:
	go build -o bin/tfgen

install: build
	sudo mv bin/tfgen /usr/local/bin
