.PHONY: build rpm clean
build:  export GOPATH = $(shell pwd)
build:  
	go build -o bin/vali src/main/main.go

rpm:
	tar zcvf vali-agent.tar.gz conf/vali-agent.conf bin/vali
	mkdir -p ./build
	./buildrpm.sh -n ./build -s vali-agent.spec
	rm -rf vali-agent.tar.gz

dev_upload:
	curl -F $$(ls -Art build/RPMS/| tail -n 1)=@build/RPMS/$$(ls -Art build/RPMS/| tail -n 1) download.nosa.me/upload

clean:
	rm -rf ./build vali-agent.tar.gz
