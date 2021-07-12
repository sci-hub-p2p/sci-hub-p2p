// Copyright 2021 Trim21<trim21.me@gmail.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

// Package convert convert bencode common case `[][]interface{}` to struct
package convert

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/pkg/errors"
)

const tagName = "tuple"

var errNotStruct = errors.New("can only scan slice in to ptr of struct")

// ScanSlice row should be omitted after passed to this function.
func ScanSlice(row []interface{}, dest interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("Unknown panic when scan row into struct")
			}
		}
	}()
	rv := reflect.ValueOf(dest)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return errNotStruct
	}
	rv = reflect.Indirect(rv)

	for i := 0; i < rv.NumField(); i++ {
		val := rv.Field(i)

		f := rv.Type().Field(i)
		v := f.Tag.Get(tagName)

		index, err := strconv.Atoi(v)
		if err != nil {
			return errors.Wrapf(err, "can't convert %s to int", v)
		}
		err = convert(f.Name, row[index], val)
		if err != nil {
			return err
		}
	}

	return err
}

func convert(name string, data interface{}, val reflect.Value) error {
	var err error
	k := getKind(val)
	switch k {
	case reflect.Bool:
		err = convertToBool(name, data, val)
	case reflect.String:
		err = convertToString(name, data, val)
	case reflect.Int:
		err = convertToInt(name, data, val)
	case reflect.Uint8:
		err = convertToUInt(name, data, val)
	case reflect.Slice:
		err = convertToSlice(name, data, val)
	default:
		return fmt.Errorf("'%s' expected type '%s', got type '%s', value: '%v'\nmaybe it's not implymented yet",
			name, val.Type(), k, data)
	}

	return err
}

func getKind(val reflect.Value) reflect.Kind {
	kind := val.Kind()

	switch {
	case kind >= reflect.Int && kind <= reflect.Int64:
		return reflect.Int
	case kind >= reflect.Uint && kind <= reflect.Uint64:
		return reflect.Uint
	case kind >= reflect.Float32 && kind <= reflect.Float64:
		return reflect.Float32
	default:
		return kind
	}
}

func convertToFloat(v interface{}) {

}

func convertToString(name string, data interface{}, val reflect.Value) error {
	dataVal := reflect.Indirect(reflect.ValueOf(data))
	dataKind := getKind(dataVal)

	switch {
	case dataKind == reflect.String:
		val.SetString(dataVal.String())
	default:
		return fmt.Errorf(
			"'%s' expected type '%s', got unconvertible type '%s', value: '%v'",
			name, val.Type(), dataVal.Type(), data)
	}

	return nil
}

func convertToSlice(name string, data interface{}, val reflect.Value) error {
	set := false
	dataVal := reflect.Indirect(reflect.ValueOf(data))
	dataValKind := dataVal.Kind()
	valType := val.Type()
	valElemType := valType.Elem()
	sliceType := reflect.SliceOf(valElemType)

	// If we have a non array/slice type then we first attempt to convert.
	if dataValKind != reflect.Array && dataValKind != reflect.Slice {
		return fmt.Errorf(
			"'%s': source data must be an array or slice, got %s", name, dataValKind)
	}

	// If the input value is nil, then don't allocate since empty != nil
	if dataVal.IsNil() {
		return nil
	}

	valSlice := val

	if valSlice.IsNil() {
		// Make a new slice to hold our result, same size as the original data.
		valSlice = reflect.MakeSlice(sliceType, dataVal.Len(), dataVal.Len())
	}

	// Accumulate any errors

	for i := 0; i < dataVal.Len(); i++ {
		currentData := dataVal.Index(i).Interface()
		for valSlice.Len() <= i {
			valSlice = reflect.Append(valSlice, reflect.Zero(valElemType))
		}

		currentField := valSlice.Index(i)
		// bytes
		if currentField.Type().Kind() == reflect.Uint8 {
			val.SetBytes(dataVal.Bytes())
			set = true
			break
		} else {
			fieldName := name + "[" + strconv.Itoa(i) + "]"
			if err := convert(fieldName, currentData, currentField); err != nil {
				return err
			}
		}
	}

	if !set {
		// Finally, set the value to the slice we built up
		val.Set(valSlice)
	}

	return nil
}

func convertToUInt(name string, data interface{}, val reflect.Value) error {
	dataVal := reflect.Indirect(reflect.ValueOf(data))
	dataKind := getKind(dataVal)

	switch dataKind {
	case reflect.Int:
		i := dataVal.Int()
		if i < 0 {
			return fmt.Errorf("cannot parse '%s', %d overflows uint",
				name, i)
		}
		val.SetUint(uint64(i))
	case reflect.Uint:
		val.SetUint(dataVal.Uint())
	default:
		return fmt.Errorf(
			"'%s' expected type '%s', got unconvertible type '%s', value: '%v'",
			name, val.Type(), dataVal.Type(), data)
	}

	return nil
}

func convertToInt(name string, data interface{}, val reflect.Value) error {
	dataVal := reflect.Indirect(reflect.ValueOf(data))
	dataKind := getKind(dataVal)

	switch dataKind {
	case reflect.Int:
		val.SetInt(dataVal.Int())
	case reflect.Uint:
		val.SetInt(int64(dataVal.Uint()))
	case reflect.Float32:
		val.SetInt(int64(dataVal.Float()))
	default:
		return fmt.Errorf(
			"'%s' expected type '%s', got unconvertible type '%s', value: '%v'",
			name, val.Type(), dataVal.Type(), data)
	}

	return nil
}

func convertToBool(name string, data interface{}, val reflect.Value) error {
	dataVal := reflect.Indirect(reflect.ValueOf(data))
	dataKind := getKind(dataVal)
	switch {
	case dataKind == reflect.Bool:
		val.SetBool(dataVal.Bool())
	default:
		return fmt.Errorf(
			"'%s' expected type '%s', got unconvertible type '%s', value: '%v'",
			name, val.Type(), dataVal.Type(), data)
	}

	return nil
}
