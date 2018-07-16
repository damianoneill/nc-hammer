package cmd_test

import (
	"reflect"
	"testing"

	. "github.com/nc-hammer/cmd"
)

/*
	NOTE: UNFINISHED
*/
func TestBuildTestSuite(t *testing.T) {

	mockPath := ""
	testType := BuildTestSuite(mockPath)

	// Check if the passed interface is a pointer
	if reflect.ValueOf(testType).Kind() != reflect.Ptr {
		t.Errorf("Testsuite is not pointer")
	}
	//if got.Kind() != reflect.Invalid {
	//	t.Errorf("Testsuite was not initialised correctly")
	//}

	got := reflect.ValueOf(testType).Elem()
	values := make([]interface{}, got.NumField())
	// iterate through the struct's fields, checking for validity at each point
	for i := 0; i < got.NumField(); i++ {
		values[i] = got.Field(i).Interface()
	}
	//fmt.Println(values) // printing out fields in struct to test later

}
