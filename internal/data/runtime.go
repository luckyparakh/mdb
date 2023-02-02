package data

import (
	"fmt"
	"log"
	"strconv"
	"strings"
)

type Runtime int32
type RuntimeErr struct {
	msg string
}

// Deliberately using a value receiver rather than a pointer receiver
// The rule about pointers vs. values for receivers is that value methods can be invoked on
// pointers and values, but pointer methods can only be invoked on pointers.
// Value receivers are concurrency safe, while pointer receivers are not concurrency safe.

func (r Runtime) MarshalJSON() ([]byte, error) {
	jv := strconv.Quote(fmt.Sprintf("%d mins", r))
	return []byte(jv), nil
}

func (r *Runtime) UnmarshalJSON(jsonValue []byte) error {
	iString, err := strconv.Unquote(string(jsonValue))
	if err != nil {
		return err
	}
	iStringSlice := strings.Split(iString, " ")
	if len(iStringSlice) != 2 {
		return RuntimeErr{msg: "Lenght of input should be 2"}
	}
	duration, err := strconv.ParseInt(iStringSlice[0], 10, 32)
	if err != nil {
		log.Println("Minutes should be integer")
		return RuntimeErr{msg: "Minutes should be integer"}
	}
	if iStringSlice[1] != "mins" || duration < 2 || duration > 200 {
		return RuntimeErr{msg: "Format is 'integer between 2 and 200' followed by ' mins'"}
	}
	*r = Runtime(duration)
	log.Println(*r)
	return nil
}

func (r RuntimeErr) Error() string {
	return r.msg
}
