package djinn

import "testing"

func tmplfnc() error {
	return nil
}

func TestConfigure(t *testing.T) {
	m := make(map[string]string)
	m["testingTmpl"] = "testing template"
	mm := make(map[string]interface{})
	mm["testingFn"] = tmplfnc
	c := New(CacheOn(TLRUCache(1)), Loaders(MapLoader(m)), TemplateFunctions(mm))
	c.Configure()
	if c.Cache == nil || c.cached == false {
		t.Errorf("Cache configuration not configured as expected.")
	}
	_, err := c.Fetch("testingTmpl")
	if err != nil {
		t.Errorf("Loaders Conf function not configured as expected.")
	}
	if _, ok := c.GetFuncs()["testingFn"]; !ok {
		t.Errorf("TemplateFunctions not configured as expected.")
	}
}
