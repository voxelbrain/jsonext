package jsonext

import (
	"fmt"
	"reflect"
	"testing"
)

func TestUnmarshalFlat(t *testing.T) {
	type Thing struct {
		ID         string `json:"_id"`
		DeleteFlag bool   `json:"_delete"`
		CatchAll   `jsonext:"catchall"`
	}
	json := `
	{
		"_id": "abc",
		"_delete": true,
		"tags": [
			"some",
			"tags"
		],
		"stuff": {
			"some_stuff": 1,
			"more_stuff": 2
		}
	}`
	expected := Thing{
		ID:         "abc",
		DeleteFlag: true,
		CatchAll: CatchAll{
			"tags": []interface{}{"some", "tags"},
			"stuff": map[string]interface{}{
				"some_stuff": float64(1),
				"more_stuff": float64(2),
			},
		},
	}

	thing := Thing{}

	err := Unmarshal([]byte(json), &thing)
	if err != nil {
		t.Fatalf("Could not unmarshal: %s", err)
	}
	if !reflect.DeepEqual(thing, expected) {
		t.Fatalf("Unexpected result value %#v, expected %#v", thing, expected)
	}
}

func TestUnmarshalDeep(t *testing.T) {
	type SubThing struct {
		ID       string `json:"_id"`
		CatchAll `jsonext:"catchall"`
	}
	type Thing struct {
		ID       string   `json:"_id"`
		Thing    SubThing `json:"thing" jsonext:"descend"`
		CatchAll `jsonext:"catchall"`
	}
	json := `
	{
		"_id": "abc",
		"thing": {
			"_id": "def",
			"stuff": ["some", "stuff"]
		},
		"stuff": {
			"some_stuff": 1,
			"more_stuff": 2
		}
	}`
	expected := Thing{
		ID: "abc",
		Thing: SubThing{
			ID: "def",
			CatchAll: CatchAll{
				"stuff": []interface{}{"some", "stuff"},
			},
		},
		CatchAll: CatchAll{
			"stuff": map[string]interface{}{
				"some_stuff": float64(1),
				"more_stuff": float64(2),
			},
		},
	}

	thing := Thing{}

	err := Unmarshal([]byte(json), &thing)
	if err != nil {
		t.Fatalf("Could not unmarshal: %s", err)
	}
	if !reflect.DeepEqual(thing, expected) {
		t.Fatalf("Unexpected result value %#v expected %#v", thing, expected)
	}
}

func TestUnmarshalSlice(t *testing.T) {
	type Thing struct {
		A     string   `json:"a"`
		Other CatchAll `jsonext:"catchall"`
	}

	json := `
	[
		{
			"a": "1",
			"b": "2",
			"c": "3"
		},

		{
			"a": "4",
			"prop": "val"
		}
	]`

	expected := []Thing{
		Thing{
			A: "1",
			Other: CatchAll{
				"b": "2",
				"c": "3",
			},
		},
		Thing{
			A: "4",
			Other: CatchAll{
				"prop": "val",
			},
		},
	}

	var things []Thing

	err := Unmarshal([]byte(json), &things)
	if err != nil {
		t.Fatalf("Could not unmarshal: %s", err)
	}
	if !reflect.DeepEqual(things, expected) {
		t.Fatalf("Unexpected result value %#v expected %#v", things, expected)
	}
}

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
