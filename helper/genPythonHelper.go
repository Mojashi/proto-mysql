package helper

import (
	"github.com/Mojashi/proto-mysql/dep"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
	"google.golang.org/protobuf/proto"
)

func genPythonHelper(dep dep.INameSpace, f *descriptor.FileDescriptorProto) []*plugin.CodeGeneratorResponse_File {
	return []*plugin.CodeGeneratorResponse_File{
		&plugin.CodeGeneratorResponse_File{
			Name: proto.String(f.GetName() + "_helper.py"),
			Content: proto.String(`
from typing import Any,Mapping,List

def getColumnNames() -> List[str]:
	return "%s"

# convert proto message class variable to INSERT-ready dictionary
def convProtoClassToData() -> Mapping[str, Any]:
	return {}
			`),
		},
	}
}
