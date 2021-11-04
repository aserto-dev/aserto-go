package pbutil

import (
	"encoding/json"
	"errors"
	"strings"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/structpb"
)

var ErrBadMask = errors.New("invalid mask")

func Select(msg proto.Message, paths ...string) (*structpb.Struct, error) {
	mask, err := newFieldMask(msg, paths...)
	if err != nil {
		return nil, err
	}

	msgStruct := emptyStruct()
	for _, path := range mask.GetPaths() {
		if err := addMessagePathToStruct(msg, path, msgStruct); err != nil {
			return nil, err
		}
	}

	return msgStruct, nil
}

func emptyStruct() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]interface{}{})
	return s
}

func messageStruct(msg proto.Message) (*structpb.Struct, error) {
	jsonMsg, err := protojson.Marshal(msg)
	if err != nil {
		return nil, err
	}

	var mapMsg map[string]interface{}
	if err := json.Unmarshal(jsonMsg, &mapMsg); err != nil {
		return nil, err
	}

	return structpb.NewStruct(mapMsg)
}

func fieldValueToStructValue(msg protoreflect.Message, fieldName string) (*structpb.Value, error) {
	field := msg.Descriptor().Fields().ByTextName(fieldName)
	value := msg.Get(field)

	if field.Kind() != protoreflect.MessageKind {
		msgVal := value.Interface()
		return structpb.NewValue(msgVal)
	}

	structValue, err := messageStruct(value.Message().Interface())
	if err != nil {
		return nil, err
	}

	return structpb.NewStructValue(structValue), nil
}

func newFieldMask(msg proto.Message, paths ...string) (*fieldmaskpb.FieldMask, error) {
	mask, err := fieldmaskpb.New(msg, paths...)
	if err != nil {
		return nil, err
	}

	mask.Normalize()

	if !mask.IsValid(msg) {
		return nil, ErrBadMask
	}

	return mask, nil
}

func addMessagePathToStruct(msg proto.Message, path string, strct *structpb.Struct) error {
	parts := strings.Split(path, ".")

	leafStruct := strct
	leafMessage := msg.ProtoReflect()

	for _, part := range parts[:len(parts)-1] {
		if _, ok := leafStruct.Fields[part]; !ok {
			leafStruct.Fields[part] = structpb.NewStructValue(emptyStruct())
		}

		leafStruct = leafStruct.Fields[part].GetStructValue()
		leafMessage = leafMessage.Get(leafMessage.Descriptor().Fields().ByTextName(part)).Message()
	}

	lastPart := parts[len(parts)-1]

	structValue, err := fieldValueToStructValue(leafMessage, lastPart)
	if err != nil {
		return err
	}

	leafStruct.Fields[lastPart] = structValue

	return nil
}
