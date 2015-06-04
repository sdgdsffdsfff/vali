.PHONY: build rpm clean
build:  export GOPATH = $(shell pwd)
build:  
	go build -o bin/vali --race src/main/main.go

rpm:
	tar zcvf vali-agent.tar.gz conf/vali-agent.conf bin/vali
	mkdir -p ./build
	./buildrpm.sh -n ./build -s vali-agent.spec
	rm -rf vali-agent.tar.gz

clean:
	rm -rf ./build vali-agent.tar.gz
