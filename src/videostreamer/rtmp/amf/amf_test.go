package amf

import (
	"testing"
	"bytes"
	"reflect"
)

func deepEqual(res, val interface{}) (equal bool) {
	switch reflect.ValueOf(res).Kind() {
	case reflect.Float64:
		equal = res == reflect.ValueOf(val).Convert(reflect.TypeOf(float64(0))).Float()
	case reflect.Bool:
		equal = res == val
	case reflect.String:
		equal = res == val
	case reflect.Slice:
		valval := reflect.ValueOf(val)
		resval := reflect.ValueOf(res)
		equal = valval.Len() == resval.Len()
		if (equal) {
			for i := 0; i < valval.Len(); i++ {
				equal = deepEqual(resval.Index(i).Interface(), valval.Index(i).Interface())
				if (!equal) {
					break
				}
			}
		}
	case reflect.Map:
		valval := reflect.ValueOf(val)
		resval := reflect.ValueOf(res)
		switch reflect.ValueOf(val).Kind() {
		case reflect.Struct:
			equal = len(resval.MapKeys()) == valval.NumField()
			if (equal) {
				for _, key := range resval.MapKeys() {
					fldname := key.String()
					if !valval.FieldByName(fldname).IsValid() {
						t := reflect.TypeOf(val)
						for i := 0; i < t.NumField(); i++ {
							if (t.Field(i).Tag.Get("name") == key.String()) {
								fldname = t.Field(i).Name
								break
							}
						}
					}
					equal = deepEqual(resval.MapIndex(key).Interface(), valval.FieldByName(fldname).Interface())
					if (!equal) {
						break
					}
				}
			}
		case reflect.Map:
			equal = len(resval.MapKeys()) == len(valval.MapKeys())
			if (equal) {
				for _, key := range resval.MapKeys() {
					equal = deepEqual(resval.MapIndex(key).Interface(), valval.MapIndex(key).Interface())
					if (!equal) {
						break
					}
				}
			}
		}
	case reflect.Invalid:
		equal = val == nil
	}
	return
}

func maketest(val interface{}, expectsiz int, t *testing.T) {
	var buf bytes.Buffer
	siz,err := EncodeAMF(&buf, val)
	if err != nil {
		t.Error("err(%s) != nil", err)
	}

	if (expectsiz > 0 && expectsiz != siz) {
		t.Errorf("siz(%d) != %d", siz, expectsiz)
	}

	res, err := DecodeAMF(&buf)
	if err != nil {
		return
	}

	if !deepEqual(res, val) {
		t.Errorf("values %v(of type %T) and %v(of type %T) are not equal", res, res, val, val)
	}
}

func TestVerbatim(t *testing.T) {
	maketest(123, 9, t)
	maketest(123.0, 9, t)
	maketest(true, 2, t)
	maketest(false, 2, t)
	maketest([]int{1,2,3}, 32, t)
	maketest(map[string]AMFValue{"avc": 3, "ggg":42}, 29, t)
	maketest(struct{A string}{A: "ac"}, 12, t)
	maketest(struct{A string `name:"drrr"`}{A: "ac"}, 15, t)
	maketest(nil, 1, t)
}