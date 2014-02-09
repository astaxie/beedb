package beedb

import (
	"errors"
	"reflect"
	"strconv"
	"strings"
	"time"
)

func getTypeName(obj interface{}) (typestr string) {
	typ := reflect.TypeOf(obj)
	typestr = typ.String()

	lastDotIndex := strings.LastIndex(typestr, ".")
	if lastDotIndex != -1 {
		typestr = typestr[lastDotIndex+1:]
	}

	return
}

func snakeCasedName(name string) string {
	newstr := make([]rune, 0)
	firstTime := true

	for _, chr := range name {
		if isUpper := 'A' <= chr && chr <= 'Z'; isUpper {
			if firstTime == true {
				firstTime = false
			} else {
				newstr = append(newstr, '_')
			}
			chr -= ('A' - 'a')
		}
		newstr = append(newstr, chr)
	}

	return string(newstr)
}

func titleCasedName(name string) string {
	newstr := make([]rune, 0)
	upNextChar := true

	for _, chr := range name {
		switch {
		case upNextChar:
			upNextChar = false
			chr -= ('a' - 'A')
		case chr == '_':
			upNextChar = true
			continue
		}

		newstr = append(newstr, chr)
	}

	return string(newstr)
}

func pluralizeString(str string) string {
	if strings.HasSuffix(str, "data") {
		return str
	}
	if strings.HasSuffix(str, "y") {
		str = str[:len(str)-1] + "ie"
	}
	return str + "s"
}

func scanMapIntoStruct(obj interface{}, objMap map[string][]byte) error {
	dataStruct := reflect.Indirect(reflect.ValueOf(obj))
	if dataStruct.Kind() != reflect.Struct {
		return errors.New("expected a pointer to a struct")
	}

	for key, data := range objMap {
		structField := dataStruct.FieldByName(titleCasedName(key))
		if !structField.CanSet() {
			continue
		}

		var v interface{}

		switch structField.Type().Kind() {
		case reflect.Slice:
			v = data
		case reflect.String:
			v = string(data)
		case reflect.Bool:
			v = string(data) == "1"
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32:
			x, err := strconv.Atoi(string(data))
			if err != nil {
				return errors.New("arg " + key + " as int: " + err.Error())
			}
			v = x
		case reflect.Int64:
			x, err := strconv.ParseInt(string(data), 10, 64)
			if err != nil {
				return errors.New("arg " + key + " as int: " + err.Error())
			}
			v = x
		case reflect.Float32, reflect.Float64:
			x, err := strconv.ParseFloat(string(data), 64)
			if err != nil {
				return errors.New("arg " + key + " as float64: " + err.Error())
			}
			v = x
		case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			x, err := strconv.ParseUint(string(data), 10, 64)
			if err != nil {
				return errors.New("arg " + key + " as int: " + err.Error())
			}
			v = x
		//Now only support Time type
		case reflect.Struct:
			if structField.Type().String() != "time.Time" {
				return errors.New("unsupported struct type in Scan: " + structField.Type().String())
			}

			x, err := time.Parse("2006-01-02 15:04:05", string(data))
			if err != nil {
				x, err = time.Parse("2006-01-02 15:04:05.000 -0700", string(data))

				if err != nil {
					return errors.New("unsupported time format: " + string(data))
				}
			}

			v = x
		default:
			return errors.New("unsupported type in Scan: " + reflect.TypeOf(v).String())
		}

		structField.Set(reflect.ValueOf(v))
	}

	return nil
}

func scanStructIntoMap(obj interface{}) (map[string]interface{}, error) {
	dataStruct := reflect.Indirect(reflect.ValueOf(obj))
	if dataStruct.Kind() != reflect.Struct {
		return nil, errors.New("expected a pointer to a struct")
	}

	dataStructType := dataStruct.Type()

	mapped := make(map[string]interface{})

	for i := 0; i < dataStructType.NumField(); i++ {
		field := dataStructType.Field(i)
		fieldName := field.Name
		bb := field.Tag
		sqlTag := bb.Get("sql")
		var mapKey string
		if bb.Get("beedb") == "-" || sqlTag == "-" || reflect.ValueOf(bb).String() == "-" {
			continue
		} else if len(sqlTag) > 0 {
			sqtags := strings.Split(sqlTag, ",")
			//TODO: support tags that are common in json like omitempty
			if sqtags[0] == "-" || sqtags[0] == "" {
				continue
			}
			mapKey = sqtags[0]
		} else {
			mapKey = snakeCasedName(fieldName)
		}
		value := dataStruct.FieldByName(fieldName).Interface()

		mapped[mapKey] = value
	}

	return mapped, nil
}

func StructName(s interface{}) string {
	v := reflect.TypeOf(s)
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	return v.Name()
}

func getTableName(s interface{}) string {
	v := reflect.TypeOf(s)
	if v.Kind() == reflect.String {
		s2, _ := s.(string)
		return snakeCasedName(s2)
	}
	tn := scanTableName(s)
	if len(tn) > 0 {
		return tn
	}
	return getTableName(StructName(s))
}

func scanTableName(s interface{}) string {
	if reflect.TypeOf(reflect.Indirect(reflect.ValueOf(s)).Interface()).Kind() == reflect.Slice {
		sliceValue := reflect.Indirect(reflect.ValueOf(s))
		sliceElementType := sliceValue.Type().Elem()
		for i := 0; i < sliceElementType.NumField(); i++ {
			bb := sliceElementType.Field(i).Tag
			if len(bb.Get("tname")) > 0 {
				return bb.Get("tname")
			}
		}
	} else {
		tt := reflect.TypeOf(reflect.Indirect(reflect.ValueOf(s)).Interface())
		for i := 0; i < tt.NumField(); i++ {
			bb := tt.Field(i).Tag
			if len(bb.Get("tname")) > 0 {
				return bb.Get("tname")
			}
		}
	}
	return ""

}
