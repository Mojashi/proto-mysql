protoc-gen-mysql: main.go dep/dep.go gensql/genSQL.go helper/genPythonHelper.go gensql/mySQLOprions.pb.go
	go build -o protoc-gen-mysql

gensql/mySQLOprions.pb.go:
	protoc --go_out=. -I=. mySQLOptions.proto

.PHONY: test
test: test/test.proto protoc-gen-mysql
	protoc --plugin=protoc-gen-mysql --mysql_out=./ test/test.proto