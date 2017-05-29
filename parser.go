package envcfg

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"text/tabwriter"
)

const (
	FlagCfgFile      = "config"
	FlagCfgFileShort = "c"
	UsageFlagCfgFile = "path to config file"
)

var cfgfile string

//ParseStruct fiend tag(annotations) for each field as set value
func ParseStruct(data interface{}) error {
	if data == nil {
		return errors.New("Null refernce exception. Data equals to nil")
	}
	p, err := newParser(data)
	if err != nil {
		return err
	}
	p.configFile = newConfigFile()
	flag.StringVar(&cfgfile, FlagCfgFile, "", UsageFlagCfgFile)
	flag.StringVar(&cfgfile, FlagCfgFileShort, "", UsageFlagCfgFile)
	err = p.Init()
	if err != nil {
		return err
	}
	flag.Parse()
	if cfgfile == "" {
		//set default path(near executable)
		cfgfile, err = filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			log.Printf("error on create config file path: %s\n", err)
		}
	}
	err = p.configFile.Unmarshal(cfgfile, data)
	if err != nil {
		log.Printf("error on unmarshal config file: %s\n", err)
	}
	err = p.Parse()
	if err != nil {
		return err
	}
	return nil
}

type parser struct {
	value      reflect.Value
	rtype      reflect.Type
	configFile configFile
	parent     *parser
	childs     []*parser
	values     []*value
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
			cp.configFile = p.configFile
			err = cp.Init()
			if err != nil {
				return err
			}
			continue
		}
		//TODO: check on another type
		vl := newValue(v, p.rtype.Field(i))
		vl.owner = p
		p.values = append(p.values, vl)
	}
	if p.parent == nil {
		flag.Usage = func() { printHelp(p) }
	}
	return nil
}

func (p *parser) Parse() error {
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

func (p *parser) Path() string {
	if p.parent == nil {
		return ""
	}
	return p.parent.Path() + p.rtype.Name() + p.configFile.GroupSeparator()
}

func (p *parser) fstring(w io.Writer) {
	for _, v := range p.values {
		v.fstring(w)
	}
	for _, v := range p.childs {
		v.fstring(w)
	}
}

func printHelp(parser *parser) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', tabwriter.DiscardEmptyColumns)
	parser.fstring(w)
	fmt.Fprint(w, "\n-", FlagCfgFile, "\\-", FlagCfgFileShort, "\t", UsageFlagCfgFile)
	w.Flush()
}
