protoc-gen-mysql: main.go dep/dep.go gensql/genSQL.go helper/genPythonHelper.go gensql/mySQLOptions.pb.go
	go build -o protoc-gen-mysql

gensql/mySQLOptions.pb.go:
	protoc --go_out=. -I=. mySQLOptions.proto

.PHONY: test
test: test/test.proto protoc-gen-mysql
	protoc --plugin=protoc-gen-mysql --mysql_out=./ --python_out=./ test/test.proto &&\
	protoc --python_out=./test mySQLOptions.proto