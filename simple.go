package simpleconfig

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/kelseyhightower/envconfig"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/afero"
)

type Format string

const (
	FJSON Format = "json"
)

// Simple is a configuration builder. It assumes files at multiple paths and environment variables.
type Simple struct {
	fs afero.Fs

	name   string
	format Format
	paths  []string
}

// New creates a Simple configuration reader.
func New(name string) *Simple {
	return &Simple{
		fs: afero.NewOsFs(),

		name:   name,
		format: FJSON,
		paths:  []string{"."},
	}
}

// SetFormat sets the specific format to use.
func (s *Simple) SetFormat(f Format) {
	s.format = f
}

// AddSearchPath adds a path for Simple to search for the configuration.
// Order matters. Configs wll be read in path order.
// The current directory is assumed.
func (s *Simple) AddSearchPath(path string) {
	s.paths = append(s.paths, path)
}

// ReadIn reads configuration into the given interface.
// Config paths where a configuration file doesn't exist are ignored.
// Configs in successive paths are merged into the previous, and values overwrite.
// Environment variables take precedence over configuration files.
func (s *Simple) ReadIn(v interface{}) error {
	var datas []map[string]interface{}
	for _, path := range s.paths {
		data, err := s.readInFilePath(path)
		if err != nil {
			return fmt.Errorf("read in: %w", err)
		}
		datas = append(datas, data)
	}
	data := merge(datas...)
	if err := mapstructure.Decode(data, v); err != nil {
		return fmt.Errorf("mapstructure: %w", err)
	}
	if err := envconfig.Process(s.name, v); err != nil {
		return fmt.Errorf("envconfig: %w", err)
	}
	return nil
}

func (s *Simple) readInFilePath(path string) (map[string]interface{}, error) {
	var data map[string]interface{}
	f, err := s.fs.Open(filepath.Join(path, fmt.Sprintf("%s.%s", s.name, s.format)))
	if errors.Is(err, os.ErrNotExist) {
		return data, nil
	} else if err != nil {
		return data, fmt.Errorf("open: %w", err)
	}
	defer f.Close()
	if err := s.decode(f, &data); err != nil {
		return data, fmt.Errorf("decode: %w", err)
	}
	return data, nil
}

func (s *Simple) decode(r io.Reader, v interface{}) error {
	switch s.format {
	case FJSON:
		return json.NewDecoder(r).Decode(v)
	default:
		return errors.New("invalid format")
	}
}
