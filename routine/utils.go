package main

import "fmt"

func extendError(err error, msg string) error {
	return fmt.Errorf(msg+": %v", err)
}
