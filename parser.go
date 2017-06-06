package envcfg

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"reflect"
	"text/tabwriter"

	"github.com/maytechnet/logger"
)

const (
	FlagCfgFileShort = "c"
	FlagCfgFile      = "config"
	UsageFlagCfgFile = "path to config file or folder where can be config file"

	FlagVerboseShort = "v-cfg"
	FlagVerbose      = "verbose-cfg"
	UsageFlagVerbose = "print internal messages(info/error)"

	FlagCfgExt        = "config-ext"
	FlagCfgExtDefault = ".conf.default"
	UsageFlagCfgExt   = "config file extension to search by it"
)

var (
	cfgfile string
	verbose bool
	cfglog  logger.Logger
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
	p.configFile = newConfigFile()
	flag.StringVar(&cfgfile, FlagCfgFile, "", UsageFlagCfgFile)
	flag.StringVar(&cfgfile, FlagCfgFileShort, "", UsageFlagCfgFile)
	var ext string
	flag.StringVar(&ext, FlagCfgExt, FlagCfgExtDefault, UsageFlagCfgExt)
	//define verbose
	flag.BoolVar(&verbose, FlagVerbose, false, UsageFlagVerbose)
	flag.BoolVar(&verbose, FlagVerboseShort, false, UsageFlagVerbose)
	err = p.Init()
	if err != nil {
		return err
	}
	flag.Parse()
	logwl := "FATAL"
	if verbose {
		logwl = "DEBUG"
	}
	fmtlog := logger.NewFmtLogger(logwl)
	cfglog = fmtlog.With("enfcfg")
	cfgfile, err = configFilePath(cfgfile, ext)
	if err != nil {
		cfglog.Errorln(err)
	} else {
		err = p.configFile.Unmarshal(cfgfile, data)
		if err != nil {
			cfglog.Errorf("error on unmarshal config file: %s\n", err)
		}
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
		flag.Usage = p.usage
	}
	return nil
}

func (p *parser) Parse() error {
	for _, v := range p.values {
		err := v.define()
		if err != nil {
			return err
		}
		if verbose {
			cfglog.Debugf("key=%s;value=%v", v.Name(), v.field)
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

func (p *parser) usage() {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', tabwriter.DiscardEmptyColumns)
	p.fstring(w)
	fmt.Fprint(w, "\n-", FlagCfgFile, "\\-", FlagCfgFileShort, "\t", UsageFlagCfgFile)
	fmt.Fprintln(w)
	w.Flush()
}
