// Copyright 2017 Alter Way
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

func init() {
	Writers = make(map[string]WriterFunc)
	RegisterWriter("json", jsonWriter)
	RegisterWriter("csv", csvWriter)
}

// Writers is a map of output format name -> output function
var Writers map[string]WriterFunc

// WriterFunc is an output function for a format
type WriterFunc func(rows chan Row, out io.Writer)

// RegisterWriter adds a format writer to the map
func RegisterWriter(name string, writer WriterFunc) {
	Writers[name] = writer
}

// WriterList returns a list of all registered outputs
func WriterList() []string {
	result := make([]string, len(Writers))
	i := 0
	for key := range Writers {
		result[i] = key
		i++
	}
	return result
}

func jsonWriter(rows chan Row, out io.Writer) {
	for row := range rows {
		var (
			record map[string]string
		)
		record = row.Labels
		if record == nil {
			record = make(map[string]string)
		}
		record["__name__"] = row.Name
		record["__time__"] = fmt.Sprintf("%d", row.Time)
		record["__value__"] = fmt.Sprintf("%f", row.Value)
		marshalled, err := json.Marshal(record)
		if err == nil {
			fmt.Fprintf(out, "%s\n", string(marshalled))
		}
	}
}

func createCSVRow(row Row, columns []string) []string {
	csvRow := make([]string, 0)
	csvRow = append(csvRow, []string{
		row.Name,
		fmt.Sprintf("%d", row.Time),
		fmt.Sprintf("%f", row.Value),
	}...)
	for _, labelName := range columns[3:] {
		csvRow = append(csvRow, row.Labels[labelName])
	}
	return csvRow
}

func csvWriter(rows chan Row, out io.Writer) {
	var (
		columns []string
	)
	writer := csv.NewWriter(out)
	columns = append(columns, []string{"__name__", "__time__", "__value__"}...)
	firstRow := <-rows
	// Create a row for each included label
	for label := range firstRow.Labels {
		columns = append(columns, label)
	}
	writer.Write(columns)
	writer.Write(createCSVRow(firstRow, columns))
	for row := range rows {
		csvRow := createCSVRow(row, columns)
		writer.Write(csvRow)
	}
	writer.Flush()
}

// streamResults outputs the results in the relevant output sink
func streamResults(filepath string, format string, rows chan Row) error {
	var (
		out io.Writer
		err error
	)
	if filepath == "-" {
		out = os.Stdout
	} else {
		out, err = os.Create(filepath)
		if err != nil {
			return err
		}
	}
	Writers[format](rows, out)
	return nil
}
