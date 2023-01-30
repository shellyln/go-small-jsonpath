# Go Small JSONPath
Small, feature limited JSONPath (+dialect) implementation.

## ✅ Features
+ Query that returns a single value
+ Safe query; returns zero value on failure
+ Negative value index; index from the last element, e.g.. `foo[-1].bar`

## 🛑 Unsupported features
+ Query that returns multiple values
+ Conditional query
+ Wildcard
+ Descendant node query
+ Aggregate functions
+ Other functions

## ⭐ Dialect
### Function

#### **`first`**

Returns the first item in the array.
```js
$.foo.(first).bar
```

#### **`last`**

Returns the last item in the array.
```js
$.foo.(last).bar
```

#### **`length`**

Returns the length of the array.
```js
$.foo.(length)
```

## 🚀 Usage

```go
package main

import (
	"fmt"
	"github.com/shellyln/go-small-jsonpath/jsonpath"
)

func main() {
    json, err := jsonpath.ReadString(`{"test":[{"abc":1},{"abc":10}]}`)
    if err != nil {
        fmt.Printf("ReadString: error = %v", err)
        return
    }

    path, err := jsonpath.Compile(`$.test[1].abc`)
    if err != nil {
        fmt.Printf("Compile: error = %v", err)
        return
    }

    v, err := path.Query(json) // returns nil, float64, string, []any, map[string]any
    if err != nil {
        fmt.Printf("Query: error = %v", err)
        return
    }
    fmt.Printf("%v", v) // float64(10)

    vn := path.QueryAsNumberOrZero(json) // returns zero value on failure
    fmt.Printf("%v", vn) // float64(10)

    vs := path.QueryAsStringOrZero(json) // returns zero value on failure
    fmt.Printf("%v", vs) // "" (zero value)
}
```

## 🪄 Query examples

Data:
```json
{
    "foo": [11, 12, 13],
    "bar": "abcdefg",
    "baz": -9876,
    "qux": {
        "quux": null
    }
}
```

Queries:
```go
> $.foo[0]
11

> $['foo'][1]
12

> $['foo'][3]
error (index out of range)

> $['foo'][-1]
13

> $["foo"].(length)
3

> $["foo"].(first)
11

> $["foo"].(last)
13

> $.bar
"abcdefg"

> $.bar.(length)
error (unsupported function operand)

> $.qux.quux
nil

> $['qux'].quux
nil

> $['qux']["quux"]
nil

> $.foo
map[string]any{...}

> $.qux
[]any{...}
```

## ⚖️ License

MIT  
Copyright (c) 2023 Shellyl_N and Authors.
