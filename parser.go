package envcfg

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"reflect"
	"text/tabwriter"
)

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
		err := v.define()
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

func (p *parser) fstring(w io.Writer) {
	if len(p.values) > 0 {
		fmt.Fprintln(w, "-----\t", "-------\t", "-----\t", "---------\t", "--------\t", "------------\t")
	}
	for _, v := range p.values {
		v.fstring(w)
	}
	for _, v := range p.childs {
		v.fstring(w)
	}
}

func printHelp(parser *parser) {
	w := tabwriter.NewWriter(os.Stdout, 3, 3, 3, ' ', tabwriter.AlignRight)
	fmt.Fprintln(w, "Flag:\t", "EnvVar:\t", "Type:\t", "Required:\t", "Default:\t", "Description:\t")
	parser.fstring(w)
	w.Flush()
}
