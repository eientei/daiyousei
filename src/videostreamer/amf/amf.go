package amf

import (
	"io"
	"videostreamer/binutil"
	"videostreamer/check"
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

func writeType(out io.Writer, value int) {
	binutil.WriteInt(out, value, 1)
}

func EncodeAMF(out io.Writer, plain AMFValue) (err error) {
	defer check.CheckPanicHandler(&err)

	value := reflect.ValueOf(plain)

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
		writeType(out, AMF_NUMBER)
		binutil.WriteDouble64(out, value.Convert(reflect.TypeOf(float64(0))).Float())
	case reflect.Bool:
		writeType(out, AMF_BOOL)
		binutil.WriteBoolean(out, value.Bool())
	case reflect.String:
		writeType(out, AMF_STRING)
		binutil.WriteInt(out, len(value.String()), 2)
		binutil.WriteString(out, value.String())
	case reflect.Invalid:
		writeType(out, AMF_NULL)
	case reflect.Slice:
		writeType(out, AMF_ARRAY)
		binutil.WriteInt(out, value.Len(), 4)
		for i := 0; i < value.Len(); i++ {
			check.Check0(EncodeAMF(out, value.Index(i).Interface()))
		}
	case reflect.Map:
		writeType(out, AMF_MAP)
		for _, v := range value.MapKeys() {
			binutil.WriteInt(out, v.Len(), 2)
			binutil.WriteString(out, v.String())
			check.Check0(EncodeAMF(out, value.MapIndex(v).Interface()))
		}
		binutil.WriteInt(out, AMF_END, 3)
	case reflect.Struct:
		writeType(out, AMF_OBJECT)
		typ := value.Type()
		for i := 0; i < value.NumField(); i++ {
			fld := typ.Field(i)
			tagname := fld.Tag.Get("name")
			fldname := fld.Name
			name := fldname
			if tagname != "" {
				name = tagname
			}

			binutil.WriteInt(out, len(name), 2)
			binutil.WriteString(out, name)
			check.Check0(EncodeAMF(out, value.FieldByName(fldname).Interface()))
		}
		binutil.WriteInt(out, AMF_END, 3)
	default:
		writeType(out, AMF_UNDEFINED)
	}

	return
}

func DecodeAMF(in io.Reader) (ret AMFValue, err error) {
	defer check.CheckPanicHandler(&err)

	var siz int
	ret = binutil.ReadInt(in, 1)
	switch ret {
	case AMF_NUMBER:
		ret = binutil.ReadDobule64(in)
	case AMF_BOOL:
		ret = binutil.ReadInt(in, 1) != 0
	case AMF_STRING:
		siz = binutil.ReadInt(in, 2)
		ret = binutil.ReadString(in, siz)
	case AMF_ARRAY:
		siz = binutil.ReadInt(in, 4)
		arr := make(AMFArray, siz)
		for i := 0; i < siz; i++ {
			arr[i] = check.Check1(DecodeAMF(in))
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
			siz = binutil.ReadInt(in, 2)
			str = binutil.ReadString(in, siz)
			ret = check.Check1(DecodeAMF(in))
			if siz == 0 && ret == AMF_END {
				break
			}
			rmap[str] = ret
		}
		ret = rmap
	}
	return
}
