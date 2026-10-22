//go:build !windows

package graph

import (
	"errors"
	"os"
	"os/user"
	"reflect"
	"strconv"
)

func privateGTDCurrentIdentity() (string, error) {
	account, err := user.Current()
	if err != nil {
		return "", err
	}
	return account.Uid, nil
}

func privateGTDOwnerIdentity(path string) (string, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return "", err
	}
	value := reflect.Indirect(reflect.ValueOf(info.Sys()))
	if !value.IsValid() || value.Kind() != reflect.Struct {
		return "", errors.New("gtd projection: owner_unavailable")
	}
	field := value.FieldByName("Uid")
	if !field.IsValid() || !field.CanUint() {
		return "", errors.New("gtd projection: owner_unavailable")
	}
	return strconv.FormatUint(field.Uint(), 10), nil
}
