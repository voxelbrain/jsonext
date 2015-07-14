package jsonext

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"
)

// CatchAll is the expected type of the field marked with the
// `catchall` tag.
type CatchAll map[string]interface{}

var catchAllType = reflect.TypeOf(CatchAll{})

// Unmarshal behaves exactly like encoding/json.Unmarhsal.
//
// Additionally, though, Unmarshal supports the `jsonexp` tag with
// 2 possible values.
// If a field of the type CatchAll has the tag `catchall`, every JSON
// field which could not be mapped to a struct member will be put in the
// CatchAll field.
// If a field of a struct type has the tag `descend`, jsonexp will
// recurse into the struct and look for a nested CatchAll field. If
// the `descend` tag is not set on a struct member, the normal JSON
// unmarshaller will be called.
func Unmarshal(data []byte, v interface{}) error {
	return NewDecoder(bytes.NewReader(data)).Decode(v)
}

type Decoder struct {
	*json.Decoder
}

func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{json.NewDecoder(r)}
}

func (d *Decoder) Decode(v interface{}) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return &json.InvalidUnmarshalError{reflect.TypeOf(v)}
	}

	if rv.Elem().Kind() == reflect.Struct {
		return d.decodeStruct(rv)
	}

	if rv.Elem().Kind() == reflect.Slice {
		return d.decodeSlice(rv)
	}

	return d.Decoder.Decode(v)
}

func (d *Decoder) decodeStruct(rv reflect.Value) error {
	var data map[string]interface{}
	err := d.Decoder.Decode(&data)
	if err != nil {
		return err
	}
	return d.descendStruct(rv.Elem(), data)
}

func (d *Decoder) decodeSlice(rv reflect.Value) error {
	var list []json.RawMessage
	d.Decoder.Decode(&list)

	e := rv.Elem()
	t := e.Type().Elem()

	for _, raw := range list {
		v := reflect.New(t)
		var data map[string]interface{}

		if err := json.Unmarshal(raw, &data); err != nil {
			return err
		}
		if err := d.descendStruct(v.Elem(), data); err != nil {
			return err
		}
		e = reflect.Append(e, v.Elem())
	}
	rv.Elem().Set(e)
	return nil
}

func (d *Decoder) descendStruct(rv reflect.Value, data map[string]interface{}) error {
	if data == nil {
		return nil
	}
	t := rv.Type()

	var rca reflect.Value
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldv := rv.Field(i)
		jsonFieldname := jsonFieldname(field)
		tag := field.Tag.Get("jsonext")
		switch tag {
		case "descend":
			if field.Type.Kind() != reflect.Struct {
				return fmt.Errorf("Cannot descend into field %s, because it is not a struct", field.Name)
			}
			if jsonFieldname == "" || jsonFieldname == "-" {
				break
			}
			subData := data[jsonFieldname]
			if subData == nil {
				break
			}
			err := d.descendStruct(fieldv, subData.(map[string]interface{}))
			if err != nil {
				return err
			}
			delete(data, jsonFieldname)
		case "catchall":
			if field.Type != catchAllType {
				return fmt.Errorf("Field %s has tag catchall but does not have type CatchAll", field.Name)
			}
			rca = fieldv
		case "":
			err := remarshal(fieldv.Addr().Interface(), data[jsonFieldname])
			if err != nil {
				return fmt.Errorf("Value for %s did not marshal into Go type %s: %s", jsonFieldname, field.Type, err)
			}
			delete(data, jsonFieldname)
		default:
			return fmt.Errorf("Unknown tag %s on field %s", tag, field.Name)
		}
	}

	// Data now contains only the fields which could not be
	// mapped onto struct fields.
	rca.Set(reflect.ValueOf(data))

	return nil
}

func jsonFieldname(f reflect.StructField) string {
	jsonTag := strings.Split(f.Tag.Get("json"), ",")[0]
	if jsonTag == "" {
		return f.Name
	}
	return jsonTag
}

func remarshal(dst interface{}, src interface{}) error {
	data, err := json.Marshal(src)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dst)
}
