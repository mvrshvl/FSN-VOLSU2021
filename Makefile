build:
	go build -o /usr/local/bin/fsnservice ./cmd/service/main.go
	go build -o /usr/local/bin/fsn ./cmd/fsnotify/main.go

install: build
	cp ./cmd/service/fsn.service /lib/systemd/system/.
	chmod 755 /lib/systemd/system/fsn.service
	cd /tmp
	useradd fsn -s /sbin/nologin -M
	systemctl enable fsn.service
	systemctl start fsn

add-test:
	fsn -add $(shell pwd)/test/test.txt -on_create $(shell pwd)/test/create.sh -on_delete $(shell pwd)/test/delete.sh -on_modify $(shell pwd)/test/modify.sh

test-delete:
	rm $(shell pwd)/test/test.txt

test-create:
	echo "create" > $(shell pwd)/test/test.txt

test-modify:
	echo "modify" > $(shell pwd)/test/test.txt

add-test-directory:
	fsn -add $(shell pwd)/test/test-directory -on_create $(shell pwd)/test/create.sh -on_delete $(shell pwd)/test/delete.sh -on_modify $(shell pwd)/test/modify.sh -r -1

test-create-directory:
	mkdir $(shell pwd)/test/test-directory

test-modify-directory:
	mkdir $(shell pwd)/test/test-directory/depth1

test-delete-directory:
	rm -rf $(shell pwd)/test/test-directory
