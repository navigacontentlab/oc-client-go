package oc_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	oc "github.com/navigacontentlab/oc-client-go/v2"
)

func propShouldHaveSingleNonEmptyStringValue(t *testing.T, prop oc.Property) {
	t.Helper()

	if len(prop.Values) != 1 {
		t.Errorf("the property %q should have 1 value, has %d",
			prop.Name, len(prop.Values))

		return
	}

	if prop.Values[0].Value == "" {
		t.Errorf("the property %q should have a non empty string value",
			prop.Name)

		return
	}
}

func propShouldNotHaveNestedValues(t *testing.T, prop oc.Property) {
	t.Helper()

	for i := range prop.Values {
		if prop.Values[i].NestedProperty != nil {
			t.Errorf("the property %q should not have nested properties",
				prop.Name)

			return
		}
	}
}

func nestedValuesShouldMatch(
	t *testing.T, prop oc.Property,
	test ...func(t *testing.T, prop oc.Property),
) {
	t.Helper()

	for _, v := range prop.Values {
		if v.NestedProperty == nil {
			t.Errorf("the property %q should have nested properties",
				prop.Name)

			return
		}

		for _, np := range v.NestedProperty.Properties {
			for i := range test {
				test[i](t, np)
			}
		}
	}
}

func propShouldNotHaveStringValues(t *testing.T, prop oc.Property) {
	t.Helper()

	for i := range prop.Values {
		if prop.Values[i].Value != "" {
			t.Errorf("the property %q should not have string values",
				prop.Name)

			return
		}
	}
}

func TestClient_Properties(t *testing.T) {
	testHasEnv(t, "OC_PROP_DOCUMENT_UUID")

	client := clientFromEnvironment(t)

	docUUID := os.Getenv("OC_PROP_DOCUMENT_UUID")

	var list oc.PropertyList

	list.Append("Headline")
	list.AddProperty("ConceptRelations", "ConceptName")

	props, err := client.Properties(context.Background(), docUUID, list)
	if err != nil {
		t.Fatal(err)
	}

	for _, prop := range props.Properties {
		switch prop.Name {
		case "Headline":
			propShouldHaveSingleNonEmptyStringValue(t, prop)
			propShouldNotHaveNestedValues(t, prop)
		case "ConceptRelations":
			propShouldNotHaveStringValues(t, prop)
			nestedValuesShouldMatch(
				t, prop, propShouldHaveSingleNonEmptyStringValue,
			)
		}
	}
}

func ExamplePropertyList_MarshalText() {
	var list oc.PropertyList

	list.Append("UUID", "Headline")
	list.AddProperty("Image", "Width", "Height")

	list.AddProperty("Author", "Title", "UUID", "URL")
	list.Ensure("Author", "Avatar").AddNested("Width", "Height")

	txt, err := list.MarshalText()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(txt))

	// Output:
	// UUID,Headline,Image[Width,Height],Author[Title,UUID,URL,Avatar[Width,Height]]
}

func ExamplePropertyList_UnmarshalText() {
	var list oc.PropertyList

	err := list.UnmarshalText([]byte("UUID,Headline,Image[Width,Height],Author[Title,UUID,URL,Avatar[Width,Height]]"))
	if err != nil {
		log.Fatal(err)
	}

	txt, err := list.MarshalText()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(txt))

	// Output:
	// UUID,Headline,Image[Width,Height],Author[Title,UUID,URL,Avatar[Width,Height]]
}

func TestPropertyList_MarshalText(t *testing.T) {
	var list oc.PropertyList

	list.AddProperty("I am illegal")

	_, err := list.MarshalText()
	if err == nil {
		t.Error("marshalling of a property with spaces should fail")
	}

	list = oc.PropertyList{}
	list.AddProperty("No[Good]")

	_, err = list.MarshalText()
	if err == nil {
		t.Error("marshalling of a property with brackets should fail")
	}
}
