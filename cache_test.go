package djinn

import (
	"html/template"
	"testing"
)

var getTests = []struct {
	name       string
	keyToAdd   string
	keyToGet   string
	expectedOk bool
}{
	{"hit", "testing/template/1", "testing/template/1", true},
	{"miss", "testing/template/1", "testing/template/missing", false},
}

func TestGet(t *testing.T) {
	t1, _ := template.New("testing/template/1").Parse(`{{define "T"}}Hello, {{.}}!{{end}}`)
	for _, tt := range getTests {
		lru := NewTLRUCache(0)
		lru.Add(tt.keyToAdd, t1)
		val, ok := lru.Get(tt.keyToGet)
		if ok != tt.expectedOk {
			t.Fatalf("%s: cache hit = %v; want %v", tt.name, ok, !ok)
		} else if ok && val != t1 {
			t.Fatalf("%s expected get to return template t1 but got %v", tt.name, val)
		}
	}
}

func TestRemove(t *testing.T) {
	t2, _ := template.New("testing/template/2").Parse(`{{define "T"}}Hello, {{.}}!{{end}}`)
	lru := NewTLRUCache(0)
	lru.Add("t2Key", t2)
	if val, ok := lru.Get("t2Key"); !ok {
		t.Fatal("TestRemove returned no match")
	} else if val != t2 {
		t.Fatalf("TestRemove failed. Expected %d, got %v", t2, val)
	}
	lru.Remove("t2Key")
	if _, ok := lru.Get("t2Key"); ok {
		t.Fatal("TestRemove returned a removed entry")
	}
}
