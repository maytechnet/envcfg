package envcfg

import (
	"flag"
	"fmt"
	"io"
	"os"
	"reflect"
	"strconv"
	"strings"
)

/*
	parse in follow importance:
		- flag
		- env variable
		- cfg file
		- default
*/

const (
	TagFlag        = "flag"
	TagEnv         = "env"
	TagDefault     = "default"
	TagRequired    = "required"
	TagDescription = "description"
	TagIgnored     = "-"
	TagNotDefined  = ""

	valIgnored    = "ignored"
	valNotDefined = "N/D"
	valDefault    = "*"
)

type value struct {
	owner    *parser
	field    reflect.Value
	tag      reflect.StructField
	flagv    string //flag name
	envv     string //env val
	cfgval   string //value from config file
	def      string //default value
	val      string //current
	required bool
	desc     string
}

func newValue(field reflect.Value, tag reflect.StructField) *value {
	v := &value{field: field, tag: tag}
	//Parse env value
	v.envv = v.tag.Tag.Get(TagEnv)
	if v.envv == TagNotDefined {
		v.envv = TagIgnored
	} else if v.envv == valDefault {
		v.envv = strings.ToUpper(v.tag.Name)
	}
	//Parse flag value
	v.flagv = tag.Tag.Get(TagFlag)
	if v.flagv == TagNotDefined {
		v.flagv = TagIgnored
	} else if strings.ToLower(v.flagv) == valDefault {
		v.flagv = strings.ToLower(tag.Name)
	}
	//Parse description
	v.desc = v.tag.Tag.Get(TagDescription)
	//Parse required
	rq := v.tag.Tag.Get(TagRequired)
	var err error
	v.required, err = strconv.ParseBool(rq)
	if err != nil {
		v.required = false
	}
	//Parse default value
	v.def = tag.Tag.Get(TagDefault)
	//set flag
	if v.flagv != TagIgnored {
		flag.StringVar(&v.val, v.flagv, "", v.desc)
	}
	return v
}

func (v *value) define() error {
	if !v.field.IsValid() {
		return v.exdef(fmt.Errorf("field:%s is invalid", v.tag.Name))
	}
	if !v.field.CanSet() {
		return v.exdef(fmt.Errorf("field:%s is not settable", v.tag.Name))
	}
	if v.field.Kind() == reflect.Struct {
		return v.exdef(fmt.Errorf("field:%s invalid, type struct is unsupported", v.tag.Name))
	}
	//check flag value
	if v.val == "" {
		//set os value
		if v.envv != "" {
			v.val = os.Getenv(v.envv)
		}
		if v.owner.configFile.Contains(v.Name()) {
			return nil //defined in config file
		}
		if v.val == "" && v.required {
			return fmt.Errorf("field:%s is required field", v.tag.Name)
		} else if v.val == "" {
			v.val = v.def
			if v.val == "" {
				return nil //default value not declared
			}
		}
	}
	switch v.tag.Type.Kind() {
	case reflect.Bool:
		i, err := strconv.ParseBool(v.val)
		if err != nil {
			return err
		}
		v.field.SetBool(i)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := strconv.ParseInt(v.val, 0, 64)
		if err != nil {
			return err
		}
		v.field.SetInt(i)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		i, err := strconv.ParseUint(v.val, 0, 64)
		if err != nil {
			return err
		}
		v.field.SetUint(i)
	case reflect.Float32, reflect.Float64:
		i, err := strconv.ParseFloat(v.val, 64)
		if err != nil {
			return err
		}
		v.field.SetFloat(i)
	case reflect.String:
		v.field.SetString(v.val)
	default:
		return v.exdef(fmt.Errorf("field:%s has unsupported type %s", v.tag.Name, v.tag.Type))
	}
	return nil
}

func (v *value) String() string {
	if v.field.Kind() == reflect.Struct || v.field.Kind() == reflect.Ptr {
		return ""
	}
	flag := v.flagv
	if flag == TagIgnored {
		flag = valIgnored
	} else {
		flag = "-" + flag
	}

	env := v.envv
	if env == TagIgnored {
		env = valIgnored
	}
	val := v.def
	if val == TagNotDefined {
		val = valNotDefined
	}
	return fmt.Sprintf("%s:\t\t\t\n\t%s\tEnvVar: %s\n\t\tRequired: %t\n\t\tDefault: %s\n\t\tDescription: %s", v.Name(), flag, env, v.required, val, v.desc)
}

func (v *value) Name() string {
	return v.owner.Path() + v.tag.Name
}

func (v *value) exdef(err error) error {
	if v.required {
		return err
	} else {
		return nil
	}
}

func (v *value) fstring(w io.Writer) {
	fmt.Fprintln(w, v.String())
}
