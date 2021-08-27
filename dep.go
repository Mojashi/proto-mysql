package main

import (
	"fmt"
	"strings"

	"github.com/golang/glog"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
)

type Path []string

type INameSpace interface {
	GetMessage(path Path) (Message, bool)
	GetEnum(path Path) (Enum, bool)
	GetNameSpace(name Path) *NameSpace
	AddEnum(enum Enum)
	AddMessage(Message Message)
	PrintTree(depth int) string
}

//impl NameSpace
type NameSpace struct {
	childNameSpaces map[string]INameSpace
	enums           map[string]Enum
	messages        map[string]Message
}

func NewNameSpace() *NameSpace {
	return &NameSpace{
		childNameSpaces: map[string]INameSpace{},
		enums:           map[string]Enum{},
		messages:        map[string]Message{},
	}
}

func (ns *NameSpace) GetMessage(path Path) (Message, bool) {
	if len(path) == 0 {
		return Message{}, false
	}
	if path[0] == "" {
		return ns.GetMessage(path[1:])
	}

	if len(path) == 1 {
		msg, ok := ns.messages[path[0]]
		return msg, ok
	} else if next, ok := ns.childNameSpaces[path[0]]; ok {
		return next.GetMessage(path[1:])
	} else {
		return Message{}, false
	}
}

func (ns *NameSpace) GetEnum(path Path) (Enum, bool) {
	if len(path) == 0 {
		return Enum{}, false
	}
	if path[0] == "" {
		return ns.GetEnum(path[1:])
	}

	if len(path) == 1 {
		enum, ok := ns.enums[path[0]]
		return enum, ok
	} else if next, ok := ns.childNameSpaces[path[0]]; ok {
		return next.GetEnum(path[1:])
	} else {
		return Enum{}, false
	}
}

//package "foo.bar" -> getNameSpace([]string{"foo", "bar"})
func (ns *NameSpace) GetNameSpace(name Path) *NameSpace {
	if len(name) == 0 {
		return ns
	}
	if name[0] == "" {
		return ns.GetNameSpace(name[1:])
	}
	var ch INameSpace
	var ok bool
	if ch, ok = ns.childNameSpaces[name[0]]; !ok {
		ch = NewNameSpace()
		ns.childNameSpaces[name[0]] = ch
	}
	return ch.GetNameSpace(name[1:])
}

func (ns *NameSpace) AddMessage(message Message) {
	ns.messages[message.message.GetName()] = message
	ns.childNameSpaces[message.message.GetName()] = &message
}
func (ns *NameSpace) AddEnum(enum Enum) {
	ns.enums[enum.enum.GetName()] = enum
}

func (ns *NameSpace) PrintTree(depth int) string {
	trees := []string{}
	for name, dep := range ns.childNameSpaces {
		trees = append(trees, strings.Repeat("\t", depth)+name+":"+dep.PrintTree(depth+1))
	}
	messages := []string{}
	for name, _ := range ns.messages {
		messages = append(messages, name)
	}
	enums := []string{}
	for name, _ := range ns.enums {
		enums = append(enums, name)
	}

	return fmt.Sprintf("messages:(%s) enums:(%s) nss:(\n%s)",
		strings.Join(messages, ","),
		strings.Join(enums, ","),
		strings.Join(trees, ",\n"),
	)
}

//impl NameSpace
type Message struct {
	NameSpace
	message *descriptor.DescriptorProto
}

func NewMessage(message *descriptor.DescriptorProto) Message {
	ret := Message{
		*NewNameSpace(),
		message,
	}
	for _, enum := range message.GetEnumType() {
		ret.AddEnum(NewEnum(enum))
	}
	for _, message := range message.GetNestedType() {
		ret.AddMessage(NewMessage(message))
	}
	return ret
}

type Enum struct {
	enum *descriptor.EnumDescriptorProto
}

func NewEnum(enum *descriptor.EnumDescriptorProto) Enum {
	return Enum{
		enum: enum,
	}
}

func analyzeDependency(req *plugin.CodeGeneratorRequest, file *descriptor.FileDescriptorProto) INameSpace {
	ns := NewNameSpace()

	files := make(map[string]*descriptor.FileDescriptorProto)
	for _, f := range req.ProtoFile {
		files[f.GetName()] = f
	}

	analyzeFile := func(f *descriptor.FileDescriptorProto) {
		cns := ns.GetNameSpace(strings.Split(f.GetPackage(), "."))
		for _, message := range f.GetMessageType() {
			cns.AddMessage(NewMessage(message))
		}
		for _, enum := range f.GetEnumType() {
			cns.AddEnum(NewEnum(enum))
		}
	}

	for _, dep := range file.Dependency {
		if f, ok := files[dep]; !ok {
			glog.Errorf("filed %s not found", f)
		} else {
			analyzeFile(f)
		}
	}
	analyzeFile(file)
	return ns
}
