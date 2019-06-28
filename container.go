package container

/*
Service container
implemented by using a hash table to store services,
holds contructor to create service and the instance created.
Hashing and collision resolution implements the same method as
python's dictionaries!
*/

import (
	"fmt"
	"hash"
	"hash/fnv"
)

type ContainerInterface interface {
	Get(s string) interface{}
	Register(s string, o interface{})
	Spawn(s string) interface{}
}

type Container struct {
	mask  uint32
	used  uint32
	table []*containerEntry
	tSize uint32
}

type createServiceFn func(c *Container) interface{}

type containerEntry struct {
	key         string
	hash        uint32
	instance    interface{}
	constructor createServiceFn
}

var fnvHash hash.Hash32 = fnv.New32a()
var perturbShift uint32 = 5

func getHash(s string) uint32 {
	fnvHash.Write([]byte(s))
	defer fnvHash.Reset()
	return fnvHash.Sum32()
}

func New(options ...uint32) *Container {
	var size uint32

	if len(options) == 0 {
		size = 64
	} else if options[0] < 6 {
		size = 8
	} else {
		size = (options[0]*3)>>1 + 2
	}

	var c Container
	c.tSize = size
	c.mask = size - 1
	c.used = 0
	c.table = make([]*containerEntry, size, size)
	return &c
}

func (c *Container) containerRealloc() {
	// reallocates table if used slots is greater than 2/3's of tSize
	if c.used > (c.tSize*682)>>10 {
		// increase table by factor of 4 times the used slots, minimum new size is 24
		newSize := c.used * 4
		for {
			if newSize >= 24 {
				break
			}

			newSize <<= 1 // double newSize
		}

		c.table = append(c.table, make([]*containerEntry, newSize-c.tSize, newSize-c.tSize)...)
		c.mask = newSize - 1
		c.tSize = newSize
	}
}

func (c *Container) Get(s string) interface{} {
	// if instance not created, creates and returns, else returns last instance created
	h := getHash(s)
	i := h
	p := h
	t := c.table

	for {
		entry := t[i&c.mask]

		if entry == nil {
			panic(fmt.Sprintf("Container could not get instance, no entry for '%s'", s))
		}

		if entry.hash == h && entry.key == s {
			if entry.instance == nil {
				entry.instance = entry.constructor(c)
			}
			return entry.instance
		}

		i = (i << 2) + i + p + 1
		p >>= perturbShift
	}
}

func (c *Container) Register(s string, o interface{}) {
	// register service in the container
	if o == nil {
		panic(fmt.Sprintf("Container could not register '%s', nil cannot be provided for service constructor or instance", s))
	}

	h := getHash(s)
	i := h
	p := h
	t := c.table

	for {
		slot := i & c.mask
		entry := t[slot]

		if entry == nil {
			new := &containerEntry{key: s, hash: h}
			switch v := o.(type) {
			case func(c *Container) interface{}:
				new.constructor = v
			default:
				new.instance = v
			}
			t[slot] = new
			c.used += 1
			c.containerRealloc() // resize table slice if neccessary
			return
		}

		if entry.hash == h && entry.key == s {
			// name already registered!!!
			panic(fmt.Sprintf("Container could not register '%s', multiple entries for single key not allowed", s))
		}

		i = (i << 2) + i + p + 1
		p >>= perturbShift
	}
}

func (c *Container) Spawn(s string) interface{} {
	// forces new instance of item and returns item, Get will now return this instance
	h := getHash(s)
	i := h
	p := h
	t := c.table

	for {
		entry := t[i&c.mask]

		if entry == nil {
			panic(fmt.Sprintf("Container could not spawn new instance, no entry for '%s'", s))
		}

		if entry.hash == h && entry.key == s {
			if entry.constructor != nil {
				entry.instance = entry.constructor(c)
			} else {
				panic(fmt.Sprintf("Container could not spawn new instance, no constuctor function provided for entry '%s'", s))
			}
			return entry.instance
		}

		i = (i << 2) + i + p + 1
		p >>= perturbShift
	}
}
