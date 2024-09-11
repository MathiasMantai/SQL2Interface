package util


import (
	"testing"
	"reflect"
	"github.com/stretchr/testify/assert"
)

func TestStringToInterfaceSlice(t *testing.T) {
	original := []string{"Test1", "Test2"}
	converted := StringToInterfaceSlice(original)

	if !reflect.DeepEqual(converted, []interface{}{"Test1", "Test2"}) {
		t.Fatalf("converted slice is not of type []interface{}")
	}
}

func TestValueInSlice(t *testing.T) {
	expected := true
	actual := ValueInSlice("Test1", StringToInterfaceSlice([]string{"Test1", "Test2"}))

	assert.Equal(t, expected, actual)
}