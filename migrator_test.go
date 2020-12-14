package migrator

import (
	"bytes"
	"context"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/certmagic"
)

type DummyStorage struct {
	stored      map[string][]byte
	lock        *sync.Mutex
	Value       string `json:"value"`
	provisioned bool
	validated   bool
}

func (ds *DummyStorage) Store(key string, value []byte) error {
	ds.stored[key] = value
	return nil
}

func (ds *DummyStorage) Load(key string) ([]byte, error) {
	val, found := ds.stored[key]
	if found {
		return val, nil
	}
	return nil, errors.New("Key not found")
}

func (ds *DummyStorage) Delete(key string) error {
	delete(ds.stored, key)
	return nil
}

func (ds *DummyStorage) Exists(key string) bool {
	_, found := ds.stored[key]
	return found
}

func (ds *DummyStorage) List(prefix string, recursive bool) ([]string, error) {
	keys := make([]string, 0, len(ds.stored))
	if prefix == "" {
		for k := range ds.stored {
			keys = append(keys, k)
		}
	}
	return keys, nil
}

func (ds *DummyStorage) Stat(key string) (certmagic.KeyInfo, error) {
	_, found := ds.stored[key]
	if found {
		return certmagic.KeyInfo{Key: key, IsTerminal: true}, nil
	}
	return certmagic.KeyInfo{Key: key}, errors.New("Key not found")
}

func (ds *DummyStorage) Lock(ctx context.Context, key string) error {
	ds.lock.Lock()
	return nil
}

func (ds *DummyStorage) Unlock(key string) error {
	ds.lock.Unlock()
	return nil
}

func (ds *DummyStorage) Provision(ctx caddy.Context) error {
	ds.provisioned = true
	return nil
}

func (ds *DummyStorage) Validate() error {
	ds.validated = true
	return nil
}

// Interface guards
var (
	_ certmagic.Storage = (*DummyStorage)(nil)
	_ caddy.Provisioner = (*DummyStorage)(nil)
	_ caddy.Validator   = (*DummyStorage)(nil)
)

func TestImport(t *testing.T) {
	storages["dummy"] = func() certmagic.Storage {
		return &DummyStorage{stored: map[string][]byte{}}
	}

	config := []byte("{\"value\": \"test json unmarshal\"}")

	s, err := InitStorage("dummy", config)
	if err != nil {
		t.Error(err)
		return
	}
	ds := s.(*DummyStorage)

	if ds.Value != "test json unmarshal" {
		t.Error("Expected value not found")
		return
	}

	if ds.provisioned == false {
		t.Error("Function Provision() not called")
		return
	}

	if ds.validated == false {
		t.Error("Function Validate() not called")
		return
	}

	caddyPath, err := ioutil.TempDir(".", "test-")
	if err != nil {
		t.Error(err)
		return
	}
	defer os.RemoveAll(caddyPath)
	f, err := ioutil.TempFile(caddyPath, "key-")
	defer f.Close()
	if err != nil {
		t.Error(err)
		return
	}
	f.Write([]byte("aaaa"))
	err = f.Sync()
	if err != nil {
		t.Error(err)
		return
	}
	err = ImportFiles(ds, caddyPath)
	if err != nil {
		t.Error(err)
		return
	}
	if len(ds.stored) == 0 {
		t.Error("File not stored")
		return
	}
}

func TestExport(t *testing.T) {
	fileContent := []byte("THIS IS A CERT")
	storages["dummy"] = func() certmagic.Storage {
		return &DummyStorage{stored: map[string][]byte{
			"cert1": fileContent,
		}}
	}

	config := []byte{}

	_, err := InitStorage("dumy", config)
	if err == nil {
		t.Errorf("Should fail. Invalid storage name")
		return
	}

	s, err := InitStorage("dummy", config)
	if err != nil {
		t.Error(err)
		return
	}
	ds := s.(*DummyStorage)

	dest, err := ioutil.TempDir(".", "test-")
	if err != nil {
		t.Error(err)
		return
	}
	defer os.RemoveAll(dest)
	ExportFiles(ds, dest)

	exported := make([]string, 0)
	filepath.Walk(dest, func(path string, info os.FileInfo, err error) error {
		exported = append(exported, path)
		return nil
	})

	if len(exported) == 0 {
		t.Errorf("Nothing exported")
		return
	}

	data, err := ioutil.ReadFile(filepath.Join(dest, "cert1"))
	if !bytes.Equal(data, fileContent) {
		t.Errorf("File content mismatch")
		return
	}
}
