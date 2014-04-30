package data

import (
	"fmt"
	"testing"
)

func TestA(t *testing.T) {
	c, err := NewConfig()
	if err != nil {
		t.Error(err)
	}
	t.Log(fmt.Sprint(c.Basic))
}
