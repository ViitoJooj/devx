package utils

import "testing"

func TestEntityNames(t *testing.T) {
	cases := map[string][2]string{
		"user": {"user", "users"}, "users": {"user", "users"},
		"component_definition":  {"component_definition", "component_definitions"},
		"component_definitions": {"component_definition", "component_definitions"},
		"component-definitions": {"component_definition", "component_definitions"},
		"category":              {"category", "categories"}, "categories": {"category", "categories"},
		"status": {"status", "statuses"}, "statuses": {"status", "statuses"},
		"address": {"address", "addresses"}, "addresses": {"address", "addresses"},
		"box": {"box", "boxes"}, "boxes": {"box", "boxes"},
		"course": {"course", "courses"}, "courses": {"course", "courses"},
		"order_item": {"order_item", "order_items"}, "key": {"key", "keys"}, "keys": {"key", "keys"},
		"class": {"class", "classes"}, "analysis": {"analysis", "analysises"},
	}
	for input, want := range cases {
		n, err := NewEntityNames(input)
		if err != nil || n.Singular != want[0] || n.Plural != want[1] {
			t.Errorf("%s: got %s/%s want %v", input, n.Singular, n.Plural, want)
		}
	}
	n, _ := NewEntityNames("component_definitions")
	t.Logf("%+v", n)
}
