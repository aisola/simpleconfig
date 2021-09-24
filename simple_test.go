package simpleconfig

import (
	"os"
	"testing"

	"github.com/matryer/is"
	"github.com/spf13/afero"
)

type Config struct {
	Test1 string `mapstructure:"test1"`
	Test2 string `mapstructure:"test2"`
	Test3 string `mapstructure:"test3" envconfig:"TEST3"`
}

func TestSimple(t *testing.T) {
	is := is.New(t)
	var config Config
	is.NoErr(os.Setenv("TEST_TEST3", "test3"))
	fs := afero.NewMemMapFs()
	is.NoErr(afero.WriteFile(fs, "./test.json", []byte(`{"test1":"test1","test2":"test1","test3":"test1"}`), 0644))
	is.NoErr(afero.WriteFile(fs, "/etc/test/test.json", []byte(`{"test2":"test2","test3":"test2"}`), 0644))
	simple := New("test")
	simple.fs = fs // override fs for tests only
	simple.SetFormat(FJSON)
	simple.AddSearchPath(".")
	simple.AddSearchPath("/etc/test")
	simple.AddSearchPath("/opt/config/test")
	is.NoErr(simple.ReadIn(&config))
	is.Equal(config.Test1, "test1")
	is.Equal(config.Test2, "test2")
	is.Equal(config.Test3, "test3")
}

func TestSimple_InvalidFormat(t *testing.T) {
	is := is.New(t)
	var config Config
	fs := afero.NewMemMapFs()
	is.NoErr(afero.WriteFile(fs, "./test.properties", []byte(`{"test1":"test1","test2":"test1","test3":"test1"}`), 0644))
	simple := New("test")
	simple.fs = fs // override fs for tests only
	simple.SetFormat("properties")
	err := simple.ReadIn(&config)
	is.True(err != nil)
	is.Equal(err.Error(), "read in: decode: invalid format")
}
