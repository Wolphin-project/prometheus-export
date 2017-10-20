package main

import (
	"bytes"
	"testing"
)

func TestCSVWriterNoLabels(t *testing.T) {
	buffer := bytes.NewBufferString("")
	row := Row{
		Name:   "a",
		Labels: nil,
		Time:   1,
		Value:  1.0,
	}
	expected := "__name__,__time__,__value__\na,1,1.000000\n"
	channel := make(chan Row, 1)
	channel <- row
	close(channel)
	csvWriter(channel, buffer)
	if buffer.String() != expected {
		t.Error("Expected:\n" + expected + "\nGot:\n" + buffer.String())
	}
}

func TestJSONWriterNoLabels(t *testing.T) {
	buffer := bytes.NewBufferString("")
	row := Row{
		Name:   "a",
		Labels: nil,
		Time:   1,
		Value:  1.0,
	}
	expected := "{\"__name__\":\"a\",\"__time__\":\"1\",\"__value__\":\"1.000000\"}\n"
	channel := make(chan Row, 1)
	channel <- row
	close(channel)
	jsonWriter(channel, buffer)
	if buffer.String() != expected {
		t.Error("Expected:\n" + expected + "\nGot:\n" + buffer.String())
	}
}
