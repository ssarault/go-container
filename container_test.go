package container

import "testing"

type mystruct struct {
	A int
	B int
}

func myrecover(t *testing.T) {
	if r := recover(); r == nil {
		t.Errorf("Register(%s) == no error for double registration, want panic", "cool")
	}
}

func TestContainer(t *testing.T) {
	c := New()
	c.Register("hello", func(c *Container) interface{} {
		m := mystruct{A: 2, B: 4}
		return &m
	})
	result := c.Get("hello").(*mystruct)
	if result.A != 2 {
		t.Errorf("Get(%s) == %d, want %d", "hello", result.A, 2)
	}
	c.Register("cool", &mystruct{A: 4, B: 0})
	result = c.Get("cool").(*mystruct)
	if result.A != 4 {
		t.Errorf("Get(%s) == %d, want %d", "cool", result.A, 4)
	}
	defer myrecover(t)
	c.Register("cool", &mystruct{A: 3, B: 0})
}
