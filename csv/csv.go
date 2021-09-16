package csv

import (
	"log"
	"reflect"
)

func ConvertToCSVformat(input []interface{}) {
	// output := make([][]string, 0)
	for _, item := range input {
		log.Println(reflect.ValueOf(item), reflect.TypeOf(item))
	}
}
