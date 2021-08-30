package gensql

import (
	"fmt"
	"strings"

	"github.com/Mojashi/proto-mysql/dep"
	"github.com/golang/glog"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/pkg/errors"
)

type MySQLDataType string

const (
	DOUBLE  MySQLDataType = "DOUBLE"
	FLOAT   MySQLDataType = "FLOAT"
	INT     MySQLDataType = "INT"
	BIGINT  MySQLDataType = "BIGINT"
	UINT    MySQLDataType = "INT UNSIGNED"
	UBIGINT MySQLDataType = "BIGINT UNSIGNED"
	ENUM    MySQLDataType = "ENUM"
	BOOLEAN MySQLDataType = "BOOLEAN"
	TEXT    MySQLDataType = "TEXT"
	BLOB    MySQLDataType = "BLOB"
	JSON    MySQLDataType = "JSON"
)

var MySQLDataTypeMap = map[descriptor.FieldDescriptorProto_Type]MySQLDataType{
	descriptor.FieldDescriptorProto_TYPE_DOUBLE:   DOUBLE,
	descriptor.FieldDescriptorProto_TYPE_FLOAT:    FLOAT,
	descriptor.FieldDescriptorProto_TYPE_INT64:    BIGINT,
	descriptor.FieldDescriptorProto_TYPE_UINT64:   UBIGINT,
	descriptor.FieldDescriptorProto_TYPE_INT32:    INT,
	descriptor.FieldDescriptorProto_TYPE_FIXED64:  UBIGINT,
	descriptor.FieldDescriptorProto_TYPE_FIXED32:  UINT,
	descriptor.FieldDescriptorProto_TYPE_BOOL:     BOOLEAN,
	descriptor.FieldDescriptorProto_TYPE_STRING:   TEXT,
	descriptor.FieldDescriptorProto_TYPE_BYTES:    BLOB,
	descriptor.FieldDescriptorProto_TYPE_UINT32:   UINT,
	descriptor.FieldDescriptorProto_TYPE_ENUM:     ENUM,
	descriptor.FieldDescriptorProto_TYPE_SFIXED32: INT,
	descriptor.FieldDescriptorProto_TYPE_SFIXED64: BIGINT,
	descriptor.FieldDescriptorProto_TYPE_SINT32:   INT,
	descriptor.FieldDescriptorProto_TYPE_SINT64:   BIGINT,
}

func enumEnum(e *descriptor.EnumDescriptorProto) (names []string) {
	vs := e.GetValue()
	names = make([]string, 0, len(vs))
	for i := 0; len(vs) > i; i++ {
		names = append(names, string(vs[i].GetName()))
	}
	return names
}

func GenMySQLDataType(dep dep.INameSpace, field *descriptor.FieldDescriptorProto) (MySQLDataType, error) {
	var mType MySQLDataType

	if field.Type != nil {
		var ok bool
		if mType, ok = MySQLDataTypeMap[field.GetType()]; !ok {
			// Message type
			mType = JSON
		}
		if mType == ENUM {
			if enum, ok := dep.GetEnum(strings.Split(field.GetTypeName(), ".")); ok {
				mType += MySQLDataType("(\"" + strings.Join(enumEnum(enum.GetEnum()), "\",\"") + "\")")
			} else {
				glog.Errorf("failed to find ENUM %s", field.GetTypeName())
				return mType, fmt.Errorf("failed to find ENUM")
			}
		}
	} else if field.TypeName != nil {
		// Message type
		mType = JSON
	} else {
		return mType, fmt.Errorf("failed to find type")
	}

	if field.GetLabel() == descriptor.FieldDescriptorProto_LABEL_REPEATED {
		// repeated type is represented as JSON
		mType = JSON
	}

	return mType, nil
}

func genColumnDefinition(dep dep.INameSpace, field *descriptor.FieldDescriptorProto) (string, error) {
	dataType, err := GenMySQLDataType(dep, field)
	nullable := "NOT NULL"
	if field.GetProto3Optional() {
		nullable = "NULL"
	}
	defaultValule := ""
	if field.DefaultValue != nil {
		defaultValule = fmt.Sprintf("DEFAULT %s", field.GetDefaultValue())
	}

	return fmt.Sprintf("%s %s %s", dataType, nullable, defaultValule), err
}

// return column definition. e.g. "id INTEGER NOT NULL"
func genCreateDefinition(dep dep.INameSpace, field *descriptor.FieldDescriptorProto) (string, error) {
	columnDefinition, err := genColumnDefinition(dep, field)
	if field.GetName() == "" {
		err = errors.Wrap(err, "field name is empty")
	}
	return fmt.Sprintf("%s %s", field.GetName(), columnDefinition), err
}

func genCreateTable(dep dep.INameSpace, mt *descriptor.DescriptorProto) string {

	createDefinitions := make([]string, 0, len(mt.Field))

	for _, field := range mt.Field {
		createDefinition, err := genCreateDefinition(dep, field)
		if err != nil {
			glog.Error(err)
			glog.Errorf("failed to process field %s in message %s", field.GetName(), mt.GetName())
		}
		createDefinitions = append(createDefinitions, "\t"+createDefinition)
	}

	createDefinitions = append(createDefinitions,
		"\tPROTO_BINARY BLOB NOT NULL",
	)

	return fmt.Sprintf("CREATE TABLE %s (\n%s\n);",
		mt.GetName(),
		strings.Join(createDefinitions, ",\n"),
	)
}

func GenSQL(dep dep.INameSpace, f *descriptor.FileDescriptorProto) string {
	createTables := make([]string, 0, len(f.MessageType))
	for _, mt := range f.MessageType {
		createTables = append(createTables, genCreateTable(dep, mt))
	}
	return strings.Join(createTables, "\n\n")
}
