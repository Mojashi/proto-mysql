protoc-gen-mysql: main.go dep.go
	go build -o protoc-gen-mysql

.PHONY: test
test: test.proto protoc-gen-mysql
	protoc --plugin=protoc-gen-mysql --mysql_out=./ test.proto