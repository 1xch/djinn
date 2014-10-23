package djinn

import "testing"

func confdjinn(cf ...Conf) *Djinn {
	c := Empty()
	c.conf = defaultconf()
	c.SetConf(cf...)
	return c
}

func tmplfnc() error {
	return nil
}

func TestCacheConfiguration(t *testing.T) {
	c := confdjinn(CacheOn(NewTLRUCache(1)))
	if c.Cache == nil || c.CacheOn == false {
		t.Errorf("CacheOn Conf function not configuring as expected.")
	}
}

func TestLoaderConfiguration(t *testing.T) {
	m := make(map[string]string)
	m["testing"] = "testing template"
	c := confdjinn(Loaders(NewMapLoader(m)))
	_, err := c.Fetch("testing")
	if err != nil {
		t.Errorf("Loaders Conf function not configuring as expected.")
	}
}

func TestFuncConfiguration(t *testing.T) {
	m := make(map[string]interface{})
	m["testing"] = tmplfnc
	c := confdjinn(TemplateFunctions(m))
	if _, ok := c.FuncMap["testing"]; !ok {
		t.Errorf("TemplateFunctions Conf function not configuring as expected.")
	}
}
