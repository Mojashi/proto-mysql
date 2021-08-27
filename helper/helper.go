package helper

import (
	"github.com/Mojashi/proto-mysql/dep"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
)

type Helper = func(dep.INameSpace, *descriptor.FileDescriptorProto) []*plugin.CodeGeneratorResponse_File

var helpers = map[string]Helper{
	"python": genPythonHelper,
}

func GetHelperGen(name string) (Helper, bool) {
	c, ok := helpers[name]
	return c, ok
}
