package tocsv

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"reflect"
)

func ObjSlice2SliceSlice(input interface{}) [][]string {
	dataValue := (reflect.ValueOf(input))
	output := make([][]string, 0)
	for i := 0; i < dataValue.Len(); i++ {
		item := reflect.Indirect(dataValue.Index(i))
		log.Println(item)

		row := make([]string, 0)
		for j := 0; j < item.NumField(); j++ {
			log.Println(item.Field(j).Interface())
			row = append(row, fmt.Sprintf("%v", item.Field(j).Interface()))
		}
		output = append(output, row)
	}
	return output
}

func WriteCsv(input [][]string, fileName string) error {
	csvFile, err := os.Create(fileName)

	if err != nil {
		log.Printf("failed creating file: %s", err)
		return err
	}

	csvwriter := csv.NewWriter(csvFile)

	for _, empRow := range input {
		_ = csvwriter.Write(empRow)
	}
	csvwriter.Flush()
	csvFile.Close()
	return nil
}
