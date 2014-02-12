package beedb

import (
	"testing"
	"time"
)

type User struct {
	SQLModel `sql:",inline"`
	Name     string `sql:"name" tname:"fn_group"`
	Auth     int    `sql:"auth"`
}

type SQLModel struct {
	Id       int       `beedb:"PK" sql:"id"`
	Created  time.Time `sql:"created"`
	Modified time.Time `sql:"modified"`
}

func TestMapToStruct(t *testing.T) {
	target := &User{}
	input := map[string][]byte{
		"name":     []byte("Test User"),
		"auth":     []byte("1"),
		"id":       []byte("1"),
		"created":  []byte("2014-01-01 10:10:10"),
		"modified": []byte("2014-01-01 10:10:10"),
	}
	err := scanMapIntoStruct(target, input)
	if err != nil {
		t.Errorf(err.Error())
	}

	_, err = scanStructIntoMap(target)

	if err != nil {
		t.Errorf(err.Error())
	}
}
