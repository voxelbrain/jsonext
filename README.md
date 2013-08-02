`jsonext` is a package augmenting Go’s core package [encoding/json].
`jsonext` was written out of the need to catch unexpected fields of a JSON
object in a map.

## Usage
`jsonext` is supposed to be a drop-in replacement for the [encoding/json]
package. The recommended usage is

```Go
import (
	// ...
	json "github.com/voxelbrain/jsonext"
	// ...
)
```

Note: The API of [encoding/json] has not been completely mirrored yet. As of now,
it is not an actual drop-in replacement.

Please see the [documentation] for details.

## Example

``` Go
func ExampleUnmarshal() {
	var jsonBlob = []byte(`{
		"Name": "Platypus",
		"Order": "Monotremata",
		"Beak": "Yellow",
		"IsAGroundhog": false
	}`)
	type Animal struct {
		Name     string
		Order    string
		CatchAll `jsonext:"catchall"`
	}
	var animal Animal
	err := Unmarshal(jsonBlob, &animal)
	if err != nil {
		fmt.Println("error:", err)
	}
	fmt.Printf("%+v", animal)
	// Output:
	// {Name:Platypus Order:Monotremata CatchAll:map[Beak:Yellow IsAGroundhog:false]}
}
```

[encoding/json]: http://godoc.org/encoding/json
[documentation]: http://godoc.org/github.com/voxelbrain/jsonext
