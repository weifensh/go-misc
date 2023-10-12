https://github.com/teh-cmc/go-internals/blob/master/chapter2_interfaces/README.md

**The `iface` structure**

`iface` is the root type that represents an interface within the runtime ([src/runtime/runtime2.go](https://github.com/golang/go/blob/bf86aec25972f3a100c3aa58a6abcbcc35bdea49/src/runtime/runtime2.go#L143-L146)).  
Its definition goes like this:
```Go
type iface struct { // 16 bytes on a 64bit arch
    tab  *itab
    data unsafe.Pointer
}
```

An interface is thus a very simple structure that maintains 2 pointers:
- `tab` holds the address of an `itab` object, which embeds the datastructures that describe both the type of the interface as well as the type of the data it points to.
- `data` is a raw (i.e. `unsafe`) pointer to the value held by the interface.


**The `itab` structure**

`itab` is defined thusly ([src/runtime/runtime2.go](https://github.com/golang/go/blob/bf86aec25972f3a100c3aa58a6abcbcc35bdea49/src/runtime/runtime2.go#L648-L658)):
```Go
type itab struct { // 40 bytes on a 64bit arch
    inter *interfacetype
    _type *_type
    hash  uint32 // copy of _type.hash. Used for type switches.
    _     [4]byte
    fun   [1]uintptr // variable sized. fun[0]==0 means _type does not implement inter.
}
```

An `itab` is the heart & brain of an interface.  

First, it embeds a `_type`, which is the internal representation of any Go type within the runtime.  
A `_type` describes every facets of a type: its name, its characteristics (e.g. size, alignment...), and to some extent, even how it behaves (e.g. comparison, hashing...)!  
In this instance, the `_type` field describes the type of the value held by the interface, i.e. the value that the `data` pointer points to.

Second, we find a pointer to an `interfacetype`, which is merely a wrapper around `_type` with some extra information that are specific to interfaces.  
As you'd expect, the `inter` field describes the type of the interface itself.

Finally, the `fun` array holds the function pointers that make up the virtual/dispatch table of the interface.  
Notice the comment that says `// variable sized`, meaning that the size with which this array is declared is *irrelevant*.  
We'll see later in this chapter that the compiler is responsible for allocating the memory that backs this array, and does so independently of the size indicated here. Likewise, the runtime always accesses this array using raw pointers, thus bounds-checking does not apply here.


### The empty interface

The datastructure for the empty interface is what you'd intuitively think it would be: an `iface` without an `itab`.  
There are two reasons for that:
1. Since the empty interface has no methods, everything related to dynamic dispatch can safely be dropped from the datastructure.
1. With the virtual table gone, the type of the empty interface itself, not to be confused with the type of the data it holds, is always the same (i.e. we talk about *the* empty interface rather than *an* empty interface).

*NOTE: Similar to the notation we used for `iface`, we'll denote the empty interface holding a type T as `eface<T>`*

`eface` is the root type that represents the empty interface within the runtime ([src/runtime/runtime2.go](https://github.com/golang/go/blob/bf86aec25972f3a100c3aa58a6abcbcc35bdea49/src/runtime/runtime2.go#L148-L151)).  
Its definition goes like this:
```Go
type eface struct { // 16 bytes on a 64bit arch
    _type *_type
    data  unsafe.Pointer
}
```
Where `_type` holds the type information of the value pointed to by `data`.  
As expected, the `itab` has been dropped entirely.

While the empty interface could just reuse the `iface` datastructure (it is a superset of `eface` after all), the runtime chooses to distinguish the two for two main reasons: space efficiency and code clarity.




### The Laws of Reflection
#### Reflection goes from interface value to reflection object.
- At the basic level, reflection is just a mechanism to examine the type and value pair stored inside an interface variable.
- Use reflect.TypeOf and reflect.ValueOf, retrieve reflect.Type and reflect.Value pieces out of an interface value. 
#### Reflection goes from reflection object to interface value.
- Given a reflect.Value we can recover an interface value using the Interface method; in effect the method packs the type and value information back into an interface representation and returns the result:
- Reflection goes from interface values to reflection objects and back again.
#### To modify a reflection object, the value must be settable.
- Just keep in mind that reflection Values need the address of something in order to modify what they represent.
