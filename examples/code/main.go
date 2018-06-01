package main

import (
	"errors"
	"os"
	"reflect"

	log "github.com/sirupsen/logrus"
)

type T struct{}

func main() {
	args := os.Args
	//	log.SetLevel(log.DebugLevel)

	function := args[1]
	log.Debug("function:" + function)

	_, err := call(function, args[2])
	if err != nil {
		log.Fatal(err.Error())
	}

}

func call(name string, params ...interface{}) (result []reflect.Value, err error) {
	t := new(T)
	f := reflect.ValueOf(t).MethodByName(name)
	if len(params) != f.Type().NumIn() {
		err = errors.New("The number of params is not adapted.")
		return nil, err
	}
	in := make([]reflect.Value, len(params))
	for k, param := range params {
		in[k] = reflect.ValueOf(param)
	}
	result = f.Call(in)
	return result, nil
}
