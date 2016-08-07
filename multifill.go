package fako

import (
	"fmt"
	"reflect"
	"strings"
)

func testAndGen(currentData interface{}, testFunc func() error) error {

	fmt.Println("Run the test")
	err := testFunc()
	if err != nil {
		return err
	}

	// evaluate what's the actual type in the parameter
	dType := reflect.TypeOf(currentData)
	fmt.Printf("Type of data is %v\n", dType)

	// if the parameter is pointer, get the actual data
	if dType.Kind() == reflect.Ptr {
		dType = dType.Elem()
		fmt.Printf("Pointer type, evaluating referenced type %v\n", dType)
	}

	if dType.Kind() == reflect.Struct {
		// so iterate over fields
		dValue := reflect.ValueOf(currentData).Elem() // this is the actual structure
		//fmt.Printf("Struct value (to be filled) %#v\n", value)
		for i := 0; i < dType.NumField(); i++ {
			fieldValue := dValue.Field(i) // this is type Value
			fieldType := fieldValue.Type()
			fmt.Printf("Field %d: type: %v, value: %v\n", i, fieldType, fieldValue)
			//fmt.Printf("Struct field type %v: %#v\n", field.Kind(), field)
			if fieldValue.Kind() == reflect.Ptr {
				// create new object
				fmt.Printf("Pointer type found, about to create new element of type %v\n", fieldType)
				newValPtr := reflect.New(fieldType.Elem())
				fieldValue.Set(newValPtr)
				// fmt.Printf("Created new element of type %v: %#v\n", newValPtr.Type(), newValPtr)
				// recurse for newly created data
				err := testAndGen(newValPtr.Interface(), testFunc)
				if err != nil {
					fmt.Printf("Error during test: %v\n", err)
					return err
				}
			} else if fieldValue.Kind() == reflect.Slice {
				// if the slice is list of pointers, get the proper type
				itemType := fieldType.Elem()

				var newValPtr reflect.Value
				fmt.Printf("Slice of type %v found, about to insert one item\n", itemType)
				if itemType.Kind() == reflect.Ptr {
					itemType = itemType.Elem()
					fmt.Printf("Pointer type, evaluating referenced type %v\n", itemType)
					// append new item
					newValPtr = reflect.New(itemType)
					fieldValue.Set(reflect.Append(fieldValue, newValPtr))
				} else {
					// slice of non-pointer
					// append new item
					newValPtr = reflect.New(itemType)
					s := reflect.ValueOf(newValPtr.Interface()).Elem()
					fieldValue.Set(reflect.Append(fieldValue, s))
				}

				// recurse for newly created data
				err := testAndGen(newValPtr.Interface(), testFunc)
				if err != nil {
					fmt.Printf("Error during test: %v\n", err)
					return err
				}
			} else if fieldValue.Kind() == reflect.Struct {
				if strings.HasPrefix(fieldType.String(), "time.") {
					// if you try to go thru time.Time, reflect panics because of unexported fields
					fmt.Println("do not dive into time package")
					continue
				}
				fmt.Printf("non-pointer struct, just recurse for %v\n", fieldType)
				err := testAndGen(fieldValue.Addr().Interface(), testFunc)
				if err != nil {
					fmt.Printf("Error during test: %v\n", err)
					return err
				}
			}
		}
	}

	return nil
}
