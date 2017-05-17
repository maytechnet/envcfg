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
	TagValue       = "default"
	TagRequired    = "required"
	TagDescription = "description"
)

type value struct {
	field reflect.Value
	tag   reflect.StructField
	val   string
}

func newValue(field reflect.Value, tag reflect.StructField) *value {
	return &value{field: field, tag: tag}
}

func (v *value) Init() {
	if v.FlagName() != "" {
		flag.StringVar(&v.val, v.FlagName(), "", v.Description())
	}
}

func (v *value) FlagName() string {
	fl := v.tag.Tag.Get(TagFlag)
	if fl == "-" {
		return ""
	}
	if fl == "" {
		return strings.ToLower(v.tag.Name)
	}
	return fl
}

//return env tag name or field name
func (v *value) EnvVariableName() string {
	fl := v.tag.Tag.Get(TagEnv)
	if fl == "-" {
		return ""
	}
	if fl == "" {
		fl = v.tag.Name
	}
	return strings.ToUpper(fl)
}

func (v *value) Description() string {
	desc := v.tag.Tag.Get(TagDescription)
	if desc == "" {
		return fmt.Sprint("Name: ", v.tag.Name)
	}
	return desc
}

func (v *value) Required() bool {
	rq := v.tag.Tag.Get(TagRequired)
	res, err := strconv.ParseBool(rq)
	if err != nil {
		return false
	}
	return res
}

func (v *value) Value() string {
	return v.tag.Tag.Get(TagValue)
}

func (v *value) define() error {
	req := v.Required()
	if !v.field.IsValid() {
		return v.exdef(req, fmt.Errorf("field:%s is invalid", v.tag.Name))
	}
	if !v.field.CanSet() {
		return v.exdef(req, fmt.Errorf("field:%s is not settable", v.tag.Name))
	}
	if v.field.Kind() == reflect.Struct {
		return v.exdef(req, fmt.Errorf("field:%s invalid, type struct is unsupported", v.tag.Name))
	}
	//check flag value
	if v.val == "" {
		//set os value
		envvar := v.EnvVariableName()
		if envvar != "" {
			v.val = os.Getenv(envvar)
		}
		if v.val == "" && req {
			return fmt.Errorf("field:%s is required field", v.tag.Name)
		} else if v.val == "" {
			v.val = v.Value() //set default value
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
		i, err := strconv.ParseInt(v.val, 10, 64)
		if err != nil {
			return err
		}
		v.field.SetInt(i)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		i, err := strconv.ParseUint(v.val, 10, 64)
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
		return v.exdef(req, fmt.Errorf("field:%s has unsupported type %s", v.tag.Name, v.tag.Type))
	}
	return nil
}

func (v *value) String() string {
	if v.field.Kind() == reflect.Struct || v.field.Kind() == reflect.Ptr {
		return ""
	}
	flag := v.FlagName()
	if flag == "" {
		flag = "ignored"
	} else {
		flag = "-" + flag
	}

	env := v.EnvVariableName()
	if env == "" {
		env = "ignored"
	}
	val := v.Value()
	if val == "" {
		val = "N/D"
	}
	return fmt.Sprintf("%s\t%s\t%s\t%t\t%s\t%s\t", flag, env, v.tag.Type, v.Required(), val, v.Description())
}

func (v *value) exdef(rq bool, err error) error {
	if rq {
		return err
	} else {
		return nil
	}
}

func (v *value) fstring(w io.Writer) {
	fmt.Fprintln(w, v.String())
}
