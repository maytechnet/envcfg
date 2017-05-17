package envcfg

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"os"
	"reflect"
)

const strFormat = "%-12s%-12s%-12s%-12v%-12s%s"

//ParseStruct fiend tag(annotations) for each field as set value
func ParseStruct(data interface{}) error {
	if data == nil {
		return errors.New("Null refernce exception. Data equals to nil")
	}
	p, err := newParser(data)
	if err != nil {
		return err
	}
	err = p.Init()
	if err != nil {
		return err
	}
	err = p.Parse()
	if err != nil {
		return err
	}
	return nil
}

func StringStruct(data interface{}) (string, error) {
	if data == nil {
		return "", errors.New("Null refernce exception. Data equals to nil")
	}
	p, err := newParser(data)
	if err != nil {
		return "", err
	}
	err = p.Init()
	if err != nil {
		return "", err
	}
	p.Parse()
	return p.String(), nil
}

type parser struct {
	value  reflect.Value
	rtype  reflect.Type
	parent *parser
	childs []*parser
	values []*value
}

func newParser(data interface{}) (*parser, error) {
	return newChildParser(nil, reflect.ValueOf(data))
}

func newChildParser(parent *parser, rvalue reflect.Value) (*parser, error) {
	p := &parser{}
	p.value = rvalue //reflect.ValueOf(data) //get reflect value
	if p.value.Kind() == reflect.Ptr {
		//check on nil
		if p.value.IsNil() {
			return nil, errors.New("Data is nil pointer")
		}
		p.value = p.value.Elem() //get value fro pointer
	}
	p.rtype = p.value.Type() //remembre type
	p.childs = make([]*parser, 0)
	p.values = make([]*value, 0)
	p.parent = parent
	return p, nil
}

func (p *parser) Init() error {
	for i := 0; i < p.value.NumField(); i++ {
		v := p.value.Field(i)
		if v.Kind() == reflect.Struct || v.Kind() == reflect.Ptr {
			cp, err := newChildParser(p, v)
			if err != nil {
				return err
			}
			p.childs = append(p.childs, cp)
			err = cp.Init()
			if err != nil {
				return err
			}
			continue
		}
		//TODO: check on another type
		vl := newValue(v, p.rtype.Field(i))
		vl.Init()
		p.values = append(p.values, vl)
	}
	if p.parent == nil {
		flag.Usage = func() { printHelp(p) }
	}
	return nil
}

func (p *parser) Parse() error {
	if p.parent == nil {
		flag.Parse()
	}
	for _, v := range p.values {
		err := v.setVariable()
		if err != nil {
			return err
		}
	}
	for _, v := range p.childs {
		err := v.Parse()
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *parser) String() string {
	var buffer bytes.Buffer
	for _, v := range p.values {
		buffer.WriteString(v.String())
		buffer.WriteString("\n")
	}
	buffer.WriteString("\n")
	for _, v := range p.childs {
		buffer.WriteString(v.String())
		buffer.WriteString("\n")
	}
	return buffer.String()
}

func printHelp(parser *parser) {
	//fmt.Fprintln(os.Stdout, "Usage:")
	// fmt.Fprintln(os.Stdout, "igonred - this field doesn't use property")
	// fmt.Fprintln(os.Stdout, "required:true - return error if field doesn't declarated")
	fmt.Fprintf(os.Stdout, strFormat, "Flag:", "EnvVar:", "Type:", "Required:", "Default:", "Description:")
	fmt.Fprintln(os.Stdout, "")
	fmt.Fprintln(os.Stdout, parser.String())
}
