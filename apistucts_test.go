package main

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestUserJsonification(t *testing.T) {
	jsonForm := `{"screen_name": "boblechef", "id": 2920819021}`
	expectedUser := &User{"boblechef", 2920819021}
	u := new(User)
	if err := json.Unmarshal([]byte(jsonForm), u); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(u, expectedUser) {
		t.Fail()
	}
}

func TestFollowerListJsonification(t *testing.T) {
	jsonForm := `{"next_cursor_str": "3333", "users":[ {"screen_name": "boblechef", "id": 2920819021}]}`
	expectedUserList := &FollowerList{"3333", []*User{&User{"boblechef", 2920819021}}}
	l := new(FollowerList)
	if err := json.Unmarshal([]byte(jsonForm), l); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(l, expectedUserList) {
		t.Fail()
	}
}
