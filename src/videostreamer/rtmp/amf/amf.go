package amf

import (
	"io"
	"videostreamer/util"
	"reflect"
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

func assure(ret int, err error) int {
	if err != nil {
		panic(err)
	}
	return ret
}

func EncodeAMF(out io.Writer, raw AMFValue) (ret int, err error) {
	value := reflect.ValueOf(raw)

	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()

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
		ret += assure(writerType(out, AMF_NUMBER))
		ret += assure(util.EncodeDouble(out, value.Convert(reflect.TypeOf(float64(0))).Float()))
	case reflect.Bool:
		ret += assure(writerType(out, AMF_BOOL))
		ret += assure(util.EncodeBoolean(out, value.Bool()))
	case reflect.String:
		ret += assure(writerType(out, AMF_STRING))
		ret += assure(util.EncodeInt(out, len(value.String()), 2))
		ret += assure(util.EncodeString(out, value.String()))
	case reflect.Invalid:
		ret += assure(writerType(out, AMF_NULL))
	case reflect.Slice:
		ret += assure(writerType(out, AMF_ARRAY))
		ret += assure(util.EncodeInt(out, value.Len(), 4))
		for i := 0; i < value.Len(); i++ {
			ret += assure(EncodeAMF(out, value.Index(i).Interface()))
		}
	case reflect.Map:
		ret += assure(writerType(out, AMF_MAP))
		for _, v := range value.MapKeys() {
			ret += assure(util.EncodeInt(out, v.Len(), 2))
			ret += assure(util.EncodeString(out, v.String()))
			ret += assure(EncodeAMF(out, value.MapIndex(v).Interface()))
		}
		util.EncodeInt(out, AMF_END, 3)
	case reflect.Struct:
		ret += assure(writerType(out, AMF_OBJECT))
		typ := value.Type()
		for i := 0; i < value.NumField(); i++ {
			fld := typ.Field(i)
			tagname := fld.Tag.Get("name")
			fldname := fld.Name
			name := fldname
			if tagname != "" {
				name = tagname
			}

			ret += assure(util.EncodeInt(out, len(name), 2))
			ret += assure(util.EncodeString(out, name))
			ret += assure(EncodeAMF(out, value.FieldByName(fldname).Interface()))
		}
		ret += assure(util.EncodeInt(out, AMF_END, 3))
	default:
		ret += assure(writerType(out, AMF_UNDEFINED))
	}

	return ret, err
}

func DecodeAMF(in io.Reader) AMFValue {
	typ := util.DecodeInt(in, 1)
	switch typ {
	case AMF_NUMBER:
		return util.DecodeDouble(in)
	case AMF_BOOL:
		return util.DecodeInt(in, 1) != 0
	case AMF_STRING:
		siz := util.DecodeInt(in, 2)
		return string(util.DecodeBuf(in, siz))
	case AMF_ARRAY:
		siz := util.DecodeInt(in, 4)
		arr := make([]AMFValue, siz)
		for i := 0; i < siz; i++ {
			arr[i] = DecodeAMF(in)
		}
		return arr
	case AMF_MAP:
		fallthrough
	case AMF_OBJECT:
		res := make(map[string]AMFValue)
		for {
			siz := util.DecodeInt(in, 2)
			str := util.DecodeString(in, siz)
			val := DecodeAMF(in)
			if siz == 0 && val == nil {
				break
			}
			res[str] = val
		}
		return res
	default:
		return nil
	}
}