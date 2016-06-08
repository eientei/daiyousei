package amf

import (
	"io"
	"reflect"
	"videostreamer/util"
)

const (
	AMF_NUMBER    = 0x00
	AMF_BOOL      = 0x01
	AMF_STRING    = 0x02
	AMF_OBJECT    = 0x03
	AMF_NULL      = 0x05
	AMF_UNDEFINED = 0x06
	AMF_MAP       = 0x08
	AMF_END       = 0x09
	AMF_ARRAY     = 0x0a
)

type AMFValue interface{}
type AMFArray []AMFValue
type AMFMap map[string]AMFValue

func writerType(out io.Writer, value int) (int, error) {
	return util.EncodeInt(out, value, 1)
}

func EncodeAMF(out io.Writer, raw AMFValue) (ret int, err error) {
	value := reflect.ValueOf(raw)

	defer recover()

	assure := func(inval int, inerr error) {
		ret += inval
		err = inerr
		if err != nil {
			panic(err)
		}
	}

	switch value.Kind() {
	case reflect.Float32:
		fallthrough
	case reflect.Float64:
		fallthrough
	case reflect.Int8:
		fallthrough
	case reflect.Int16:
		fallthrough
	case reflect.Int32:
		fallthrough
	case reflect.Int64:
		fallthrough
	case reflect.Uint8:
		fallthrough
	case reflect.Uint16:
		fallthrough
	case reflect.Uint32:
		fallthrough
	case reflect.Uint64:
		fallthrough
	case reflect.Int:
		assure(writerType(out, AMF_NUMBER))
		assure(util.EncodeDouble(out, value.Convert(reflect.TypeOf(float64(0))).Float()))
	case reflect.Bool:
		assure(writerType(out, AMF_BOOL))
		assure(util.EncodeBoolean(out, value.Bool()))
	case reflect.String:
		assure(writerType(out, AMF_STRING))
		assure(util.EncodeInt(out, len(value.String()), 2))
		assure(util.EncodeString(out, value.String()))
	case reflect.Invalid:
		assure(writerType(out, AMF_NULL))
	case reflect.Slice:
		assure(writerType(out, AMF_ARRAY))
		assure(util.EncodeInt(out, value.Len(), 4))
		for i := 0; i < value.Len(); i++ {
			assure(EncodeAMF(out, value.Index(i).Interface()))
		}
	case reflect.Map:
		assure(writerType(out, AMF_MAP))
		for _, v := range value.MapKeys() {
			assure(util.EncodeInt(out, v.Len(), 2))
			assure(util.EncodeString(out, v.String()))
			assure(EncodeAMF(out, value.MapIndex(v).Interface()))
		}
		util.EncodeInt(out, AMF_END, 3)
	case reflect.Struct:
		assure(writerType(out, AMF_OBJECT))
		typ := value.Type()
		for i := 0; i < value.NumField(); i++ {
			fld := typ.Field(i)
			tagname := fld.Tag.Get("name")
			fldname := fld.Name
			name := fldname
			if tagname != "" {
				name = tagname
			}

			assure(util.EncodeInt(out, len(name), 2))
			assure(util.EncodeString(out, name))
			assure(EncodeAMF(out, value.FieldByName(fldname).Interface()))
		}
		assure(util.EncodeInt(out, AMF_END, 3))
	default:
		assure(writerType(out, AMF_UNDEFINED))
	}

	return
}

func DecodeAMF(in io.Reader) (ret AMFValue, err error) {
	var siz int

	defer recover()

	assure := func(inval AMFValue, inerr error) AMFValue {
		ret = inval
		err = inerr
		if err != nil {
			panic(err)
		}
		return ret
	}

	assure(util.DecodeInt(in, 1))
	switch ret {
	case AMF_NUMBER:
		assure(util.DecodeDouble(in))
	case AMF_BOOL:
		ret = assure(util.DecodeInt(in, 1)) != 0
	case AMF_STRING:
		siz = assure(util.DecodeInt(in, 2)).(int)
		assure(util.DecodeString(in, siz))
	case AMF_ARRAY:
		siz = assure(util.DecodeInt(in, 4)).(int)
		arr := make(AMFArray, siz)
		for i := 0; i < siz; i++ {
			arr[i] = assure(DecodeAMF(in))
		}
		ret = arr
	case AMF_NULL:
		ret = nil
	case AMF_MAP:
		fallthrough
	case AMF_OBJECT:
		var (
			str  string
			rmap AMFMap = make(AMFMap)
		)
		for {
			siz = assure(util.DecodeInt(in, 2)).(int)
			str = assure(util.DecodeString(in, siz)).(string)
			assure(DecodeAMF(in))
			if siz == 0 && ret == AMF_END {
				break
			}
			rmap[str] = ret
		}
		ret = rmap
	}
	return
}
