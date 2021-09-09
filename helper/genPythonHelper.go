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

// namespace is seperated by "_"
func getEnumDictName(namespace string, name string) string {
	return fmt.Sprintf("ENUMDICT_%s_%s", namespace, name)
}

func genEnumDicts(dep dep.INameSpace, namespace string) []string {
	enums := dep.GetEnums()
	cur := make([]string, 0, len(enums))
	for name, enum := range enums {
		kvs := make([]string, 0, len(enum.GetEnum().Value))
		for _, v := range enum.GetEnum().Value {
			kvs = append(kvs, fmt.Sprintf("\t%d:\"%s\"", v.GetNumber(), v.GetName()))
		}
		cur = append(cur, fmt.Sprintf("%s = {\n%s\n}",
			getEnumDictName(namespace, name),
			strings.Join(kvs, ",\n"),
		))
	}
	for name, ns := range dep.GetNameSpaces() {
		cur = append(cur, genEnumDicts(ns, namespace+"_"+name)...)
	}
	return cur
}

func genMethods(dep dep.INameSpace, mdesc *descriptor.DescriptorProto) string {
	tableName := mdesc.GetName()
	elems := []string{}
	columns := []string{}

	for _, fdesc := range mdesc.Field {
		columns = append(columns, fdesc.GetName())
		name := "value." + fdesc.GetName()
		ptype := fdesc.GetType()

		switch t, _ := gensql.GenMySQLDataType(dep, fdesc); t.GetType() {
		case gensql.JSON:
			if fdesc.GetLabel() == descriptor.FieldDescriptorProto_LABEL_REPEATED &&
				ptype != descriptor.FieldDescriptorProto_TYPE_MESSAGE {
				elems = append(elems, fmt.Sprintf("json.dumps(list(%s))", name))
			} else {
				elems = append(elems, fmt.Sprintf("json_format.MessageToJson(%s)", name))
			}
		case gensql.ENUM:
			terms := strings.Split(fdesc.GetTypeName(), ".")
			elems = append(elems, fmt.Sprintf("%s[%s]",
				getEnumDictName(strings.Join(terms[:len(terms)-1], "_"), terms[len(terms)-1]),
				name))
		default:
			elems = append(elems, name)
		}
	}

	columns = append(columns, "PROTO_BINARY")
	elems = append(elems, "value.SerializeToString()")

	return fmt.Sprintf(`
def get%sColumnNames() -> List[str]:
	return "%s"
	
# convert proto message class variable to INSERT-ready dictionary
def conv%sProtoClassToData(value) -> Tuple:
	return (%s)
		`, tableName, strings.Join(columns, ","), tableName, strings.Join(elems, ","))
}

func genPythonHelper(dep dep.INameSpace, f *descriptor.FileDescriptorProto) []*plugin.CodeGeneratorResponse_File {
	methods := []string{}

	for _, mdesc := range f.MessageType {
		methods = append(methods, genMethods(dep, mdesc))
	}

	return []*plugin.CodeGeneratorResponse_File{
		{
			Name: proto.String(strings.Replace(f.GetName(), ".proto", "", -1) + "_sqlhelper.py"),
			Content: proto.String(`
from typing import Any,Mapping,List,Tuple
from google.protobuf import json_format
import json
` +
				strings.Join(genEnumDicts(dep, ""), "\n\n") +
				strings.Join(methods, "\n\n")),
		},
	}
}
