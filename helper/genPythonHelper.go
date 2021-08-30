package helper

import (
	"fmt"
	"strings"

	"github.com/Mojashi/proto-mysql/dep"
	"github.com/Mojashi/proto-mysql/gensql"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
	"google.golang.org/protobuf/proto"
)

func genMethods(dep dep.INameSpace, mdesc *descriptor.DescriptorProto) string {
	tableName := mdesc.GetName()
	elems := []string{}
	columns := []string{}

	for _, fdesc := range mdesc.Field {
		columns = append(columns, fdesc.GetName())
		name := "value." + fdesc.GetName()

		if t, _ := gensql.GenMySQLDataType(dep, fdesc); t == gensql.JSON {
			elems = append(elems, fmt.Sprintf("json_format.MessageToJson(%s)", name))
		} else {
			elems = append(elems, name)
		}
	}

	columns = append(columns, "PROTO_BINARY")
	elems = append(elems, "value.SerializeToString()")

	return fmt.Sprintf(`
def get%sColumnNames() -> List[str]:
	return "%s"
	
# convert proto message class variable to INSERT-ready dictionary
def conv%sProtoClassToData(value) :
	return (%s)
		`, tableName, strings.Join(columns, ","), tableName, strings.Join(elems, ","))
}

func genPythonHelper(dep dep.INameSpace, f *descriptor.FileDescriptorProto) []*plugin.CodeGeneratorResponse_File {
	methods := []string{}

	for _, mdesc := range f.MessageType {
		methods = append(methods, genMethods(dep, mdesc))
	}

	return []*plugin.CodeGeneratorResponse_File{
		&plugin.CodeGeneratorResponse_File{
			Name: proto.String(f.GetName() + "_helper.py"),
			Content: proto.String(`
from typing import Any,Mapping,List
from google.protobuf import json_format
` + strings.Join(methods, "\n\n")),
		},
	}
}
