protoc-gen-mysql: main.go dep/dep.go gensql/genSQL.go helper/genPythonHelper.go 
	go build -o protoc-gen-mysql

.PHONY: test
test: test/test.proto protoc-gen-mysql
	protoc --plugin=protoc-gen-mysql --mysql_out=./ test/test.proto