package oc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

type PropertyResult struct {
	ContentType string     `json:"contentType"`
	Editable    bool       `json:"editable"`
	Properties  []Property `json:"properties"`
}

type Property struct {
	Name        string          `json:"name"`
	Type        string          `json:"type"`
	MultiValued bool            `json:"multiValued"`
	ReadOnly    bool            `json:"readOnly"`
	Values      []PropertyValue `json:"values"`
}

type PropertyValue struct {
	NestedProperty *PropertyResult
	Value          string
}

func (pv *PropertyValue) UnmarshalJSON(data []byte) error {
	if data[0] == '"' {
		err := json.Unmarshal(data, &pv.Value)
		if err != nil {
			return fmt.Errorf("failed to unmarshal as string: %w", err)
		}
	} else {
		var res PropertyResult

		if err := json.Unmarshal(data, &res); err != nil {
			return fmt.Errorf("failed to unmarshal as nested property: %w", err)
		}

		pv.NestedProperty = &res
	}

	return nil
}

type PropertyList []*PropertyReference

func (pl *PropertyList) AddProperty(name string, nested ...string) *PropertyReference {
	r := NewPropertyReference(name, nested...)

	*pl = append(*pl, r)

	return r
}

func (pl *PropertyList) Append(names ...string) {
	for _, name := range names {
		r := NewPropertyReference(name)

		*pl = append(*pl, r)
	}
}

func (pl *PropertyList) Ensure(names ...string) *PropertyReference {
	var ref *PropertyReference

	for _, r := range *pl {
		if r.Name == names[0] {
			ref = r
			break
		}
	}

	if ref == nil {
		ref = pl.AddProperty(names[0])
	}

	if len(names) == 1 {
		return ref
	}

	return ref.Nested.Ensure(names[1:]...)
}

func (pl *PropertyList) UnmarshalText(text []byte) error {
	var (
		stack  []*PropertyList
		buffer []byte
	)

	p := pl

	for i := range text {
		switch text[i] {
		case ' ', '\n', '\t':
			// No need to be petty, we should be able to
			// safely ignore whitespace.
			continue
		case ',':
			if len(buffer) == 0 {
				continue
			}

			p.AddProperty(string(buffer))
			buffer = buffer[0:0]
		case '[':
			if len(buffer) == 0 {
				return fmt.Errorf(
					"error at character %d, cannot add nested properties to a property without a name",
					i)
			}

			stack = append(stack, p)
			p = &p.AddProperty(string(buffer)).Nested
			buffer = buffer[0:0]
		case ']':
			if len(stack) == 0 {
				return fmt.Errorf(
					"error at character %d, unbalanced ]", i)
			}

			if len(buffer) > 0 {
				p.AddProperty(string(buffer))
				buffer = buffer[0:0]
			}

			p = stack[len(stack)-1]
			stack = stack[0 : len(stack)-1]
		default:
			buffer = append(buffer, text[i])
		}
	}

	return nil
}

func (pl *PropertyList) MarshalText() ([]byte, error) {
	var buf bytes.Buffer

	ew := &errWriter{W: &buf}

	pl.writeTo(ew)

	if ew.Err != nil {
		return nil, ew.Err
	}

	return buf.Bytes(), nil
}

func (pl *PropertyList) writeTo(ew *errWriter) {
	for i := range *pl {
		if i > 0 {
			ew.Puts(",")
		}

		(*pl)[i].writeTo(ew)
	}
}

type PropertyReference struct {
	Name   string
	Nested PropertyList
}

func NewPropertyReference(name string, nested ...string) *PropertyReference {
	r := PropertyReference{Name: name}

	r.AddNested(nested...)

	return &r
}

func (p *PropertyReference) AddNested(nested ...string) {
	for i := range nested {
		p.Nested = append(p.Nested, &PropertyReference{Name: nested[i]})
	}
}

type errWriter struct {
	W   io.Writer
	Err error
}

func (ew *errWriter) Put(data []byte) {
	if ew.Err != nil {
		return
	}

	_, err := ew.W.Write(data)
	if err != nil {
		ew.Err = err
	}
}

func (ew *errWriter) Puts(str string) {
	ew.Put([]byte(str))
}

func (p *PropertyReference) writeTo(ew *errWriter) {
	// There are probably more rules that govern property
	// names. These are the obvious ones though.
	if strings.ContainsAny(p.Name, "[] ") {
		ew.Err = fmt.Errorf("illegal property name %q", p.Name)
		return
	}

	ew.Puts(p.Name)

	if len(p.Nested) > 0 {
		ew.Puts("[")
		p.Nested.writeTo(ew)
		ew.Puts("]")
	}
}
