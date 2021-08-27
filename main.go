package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/golang/glog"
	descriptor "github.com/golang/protobuf/protoc-gen-go/descriptor"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
)

func parseReq(r io.Reader) (*plugin.CodeGeneratorRequest, error) {
	buf, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	var req plugin.CodeGeneratorRequest
	if err = proto.Unmarshal(buf, &req); err != nil {
		return nil, err
	}
	return &req, nil
}

type MySQLDataType string

var MySQLDataTypeMap = map[descriptor.FieldDescriptorProto_Type]MySQLDataType{
	descriptor.FieldDescriptorProto_TYPE_DOUBLE:   "DOUBLE",
	descriptor.FieldDescriptorProto_TYPE_FLOAT:    "FLOAT",
	descriptor.FieldDescriptorProto_TYPE_INT64:    "BIGINT",
	descriptor.FieldDescriptorProto_TYPE_UINT64:   "BIGINT UNSIGNED",
	descriptor.FieldDescriptorProto_TYPE_INT32:    "INT",
	descriptor.FieldDescriptorProto_TYPE_FIXED64:  "BIGINT UNSIGNED",
	descriptor.FieldDescriptorProto_TYPE_FIXED32:  "INT",
	descriptor.FieldDescriptorProto_TYPE_BOOL:     "BOOLEAN",
	descriptor.FieldDescriptorProto_TYPE_STRING:   "TEXT",
	descriptor.FieldDescriptorProto_TYPE_BYTES:    "BLOB",
	descriptor.FieldDescriptorProto_TYPE_UINT32:   "INT UNSIGNED",
	descriptor.FieldDescriptorProto_TYPE_ENUM:     "ENUM",
	descriptor.FieldDescriptorProto_TYPE_SFIXED32: "INT",
	descriptor.FieldDescriptorProto_TYPE_SFIXED64: "BIGINT",
	descriptor.FieldDescriptorProto_TYPE_SINT32:   "INT",
	descriptor.FieldDescriptorProto_TYPE_SINT64:   "BIGINT",
}

func enumEnum(e *descriptor.EnumDescriptorProto) (names []string) {
	vs := e.GetValue()
	names = make([]string, 0, len(vs))
	for i := 0; len(vs) > i; i++ {
		names = append(names, string(vs[i].GetName()))
	}
	return names
}

func genMySQLDataType(dep INameSpace, field *descriptor.FieldDescriptorProto) (MySQLDataType, error) {
	var mType MySQLDataType

	if field.Type != nil {
		var ok bool
		if mType, ok = MySQLDataTypeMap[field.GetType()]; !ok {
			return mType, fmt.Errorf("type %s doesn't have corresponding type in MySQL", field.Type.String())
		}
		if mType == "ENUM" {
			if enum, ok := dep.getEnum(strings.Split(field.GetTypeName(), ".")); ok {
				mType += MySQLDataType("(\"" + strings.Join(enumEnum(enum.enum), "\",\"") + "\")")
			}
		}
	} else if field.TypeName != nil {
		return mType, fmt.Errorf("type %s doesn't have corresponding type in MySQL", *field.TypeName)
	} else {
		return mType, fmt.Errorf("failed to find type")
	}

	return mType, nil
}

func genColumnDefinition(dep INameSpace, field *descriptor.FieldDescriptorProto) (string, error) {
	dataType, err := genMySQLDataType(dep, field)
	nullable := "NOT NULL"
	if field.GetProto3Optional() {
		nullable = "NULL"
	}
	defaultValule := ""
	if field.DefaultValue != nil {
		defaultValule = fmt.Sprintf("DEFAULT %s", *field.DefaultValue)
	}

	return fmt.Sprintf("%s %s %s", dataType, nullable, defaultValule), err
}

// return column definition. e.g. "id INTEGER NOT NULL"
func genCreateDefinition(dep INameSpace, field *descriptor.FieldDescriptorProto) (string, error) {
	columnDefinition, err := genColumnDefinition(dep, field)
	if field.GetName() == "" {
		err = errors.Wrap(err, "field name is empty")
	}
	return fmt.Sprintf("%s %s", field.GetName(), columnDefinition), err
}

func genCreateTable(dep INameSpace, mt *descriptor.DescriptorProto) string {

	createDefinitions := make([]string, 0, len(mt.Field))

	for _, field := range mt.Field {
		createDefinition, err := genCreateDefinition(dep, field)
		if err != nil {
			glog.Error(err)
			glog.Errorf("failed to process field %s in message %s", field.GetName(), mt.GetName())
		}
		createDefinitions = append(createDefinitions, "\t"+createDefinition)
	}

	return fmt.Sprintf("CREATE TABLE %s (\n%s\n);",
		mt.GetName(),
		strings.Join(createDefinitions, ",\n"),
	)
}

func genSQL(dep INameSpace, f *descriptor.FileDescriptorProto) string {
	log.Print(f.Dependency)
	log.Print(f.EnumType)
	createTables := make([]string, 0, len(f.MessageType))
	for _, mt := range f.MessageType {
		createTables = append(createTables, genCreateTable(dep, mt))
	}
	return strings.Join(createTables, "\n\n")
}

func processReq(req *plugin.CodeGeneratorRequest) *plugin.CodeGeneratorResponse {
	files := make(map[string]*descriptor.FileDescriptorProto)
	for _, f := range req.ProtoFile {
		files[f.GetName()] = f
	}
	var resp plugin.CodeGeneratorResponse
	for _, fname := range req.FileToGenerate {
		f := files[fname]

		dep := analyzeDependency(req, f)

		out := fname + ".sql"
		resp.File = append(resp.File, &plugin.CodeGeneratorResponse_File{
			Name:    proto.String(out),
			Content: proto.String(genSQL(dep, f)),
		})
	}

	var SupportedFeatures = uint64(plugin.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)
	resp.SupportedFeatures = proto.Uint64(SupportedFeatures)
	return &resp
}

func emitResp(resp *plugin.CodeGeneratorResponse) error {
	buf, err := proto.Marshal(resp)
	if err != nil {
		return err
	}
	_, err = os.Stdout.Write(buf)
	return err
}

func run() error {
	req, err := parseReq(os.Stdin)
	if err != nil {
		return err
	}

	resp := processReq(req)

	return emitResp(resp)
}

func main() {
	if err := run(); err != nil {
		log.Fatalln(err)
	}
}
