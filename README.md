# PROTO-MYSQL

protobuf protoc plugin for generating ```CRAETE TABLE``` statement.

## Usage
```bash=
make # build protoc-gen-mysql

protoc --plugin=protoc-gen-mysql --mysql_out=./ test.proto
```