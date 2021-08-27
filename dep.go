package main

import (
	"strings"

	"github.com/golang/glog"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
)

type Path []string

type INameSpace interface {
	getMessage(path Path) (Message, bool)
	getEnum(path Path) (Enum, bool)
	getNameSpace(name Path) *NameSpace
	addEnum(enum Enum)
	addMessage(Message Message)
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

func (ns *NameSpace) getMessage(path Path) (Message, bool) {
	if len(path) == 0 {
		return Message{}, false
	}

	if len(path) == 1 {
		msg, ok := ns.messages[path[0]]
		return msg, ok
	} else if next, ok := ns.childNameSpaces[path[0]]; ok {
		return next.getMessage(path[1:])
	} else {
		return Message{}, false
	}
}

func (ns *NameSpace) getEnum(path Path) (Enum, bool) {
	if len(path) == 0 {
		return Enum{}, false
	}

	if len(path) == 1 {
		enum, ok := ns.enums[path[0]]
		return enum, ok
	} else if next, ok := ns.childNameSpaces[path[0]]; ok {
		return next.getEnum(path[1:])
	} else {
		return Enum{}, false
	}
}

//package "foo.bar" -> getNameSpace([]string{"foo", "bar"})
func (ns *NameSpace) getNameSpace(name Path) *NameSpace {
	if len(name) == 0 {
		return ns
	}
	var ch INameSpace
	var ok bool
	if ch, ok = ns.childNameSpaces[name[0]]; !ok {
		ch = NewNameSpace()
		ns.childNameSpaces[name[0]] = ch
	}
	return ch.getNameSpace(name[1:])
}

func (ns *NameSpace) addMessage(message Message) {
	ns.messages[message.message.GetName()] = message
	ns.childNameSpaces[message.message.GetName()] = &message
}
func (ns *NameSpace) addEnum(enum Enum) {
	ns.enums[enum.enum.GetName()] = enum
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
		ret.addEnum(NewEnum(enum))
	}
	for _, message := range message.GetNestedType() {
		ret.addMessage(NewMessage(message))
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

	for _, dep := range file.Dependency {
		if f, ok := files[dep]; !ok {
			glog.Errorf("filed %s not found", f)
		} else {
			cns := ns.getNameSpace(strings.Split(f.GetPackage(), "."))
			for _, message := range f.GetMessageType() {
				cns.addMessage(NewMessage(message))
			}
			for _, enum := range f.GetEnumType() {
				cns.addEnum(NewEnum(enum))
			}
		}
	}
	return ns
}
