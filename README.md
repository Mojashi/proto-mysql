# PROTO-MYSQL

protobuf protoc plugin for generating ```CRAETE TABLE``` statement.

~~generate python,cpp mysql interface.~~

## Type Mapping
|protocol buffer | MySQL |
|-----------|---------|
|MESSAGE| JSON|
|ENUM| ENUM|
|DOUBLE| DOUBLE|
|FLOAT| FLOAT|
|INT64| BIGINT|
|UINT64| BIGINT UNSIGNED|
|INT32| INT|
|FIXED64| BIGINT UNSIGNED|
|FIXED32| INT UNSIGNED|
|BOOL| BOOLEAN|
|STRING| TEXT|
|BYTES| BLOB|
|UINT32| INT UNSIGNED|
|ENUM| ENUM|
|SFIXED32| INT|
|SFIXED64| BIGINT|
|SINT32| INT|
|SINT64| BIGINT|

## Example
### Input
```protobuf
syntax = "proto3";
package Foo;

message SearchRequest {
  string query = 1;
  int32 page_number = 2;
  int32 result_per_page = 3;
}
message User {
  int32 id = 1;
  string username = 2;
  optional int32 age = 3; 

  enum Gender {
    MALE = 0;
    FEMALE = 1;
    OTHER = 2;
  }
  Gender sgender = 6;
  SearchRequest s = 7;
}
```
### Output
```sql
CREATE TABLE SearchRequest (
	query TEXT NOT NULL ,
	page_number INT NOT NULL ,
	result_per_page INT NOT NULL 
);

CREATE TABLE User (
	id INT NOT NULL ,
	username TEXT NOT NULL ,
	age INT NULL ,
	sgender ENUM("MALE","FEMALE","OTHER") NOT NULL ,
	s JSON NOT NULL 
);
```

## Usage
```bash=
make # build protoc-gen-mysql

protoc --plugin=protoc-gen-mysql --mysql_out=./ test.proto
```