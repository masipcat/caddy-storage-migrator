package migrator

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"unsafe"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/certmagic"

	redis "github.com/gamalan/caddy-tlsredis"
)

var storages = map[string]func() certmagic.Storage{
	"redis": func() certmagic.Storage {
		return new(redis.RedisStorage)
	},
}

func InitStorage(name string, config []byte) (interface{}, error) {
	newStorage, found := storages[name]
	if !found {
		return nil, fmt.Errorf("'%s' not found", name)
	}
	storage := newStorage().(interface{})

	if len(config) > 0 {
		err := json.Unmarshal(config, storage)
		if err != nil {
			return nil, fmt.Errorf("Couldn't unmarshal config: %v", err)
		}
	}

	if prov, ok := storage.(caddy.Provisioner); ok {
		ctx, _ := caddy.NewContext(caddy.Context{Context: context.Background()})
		cfg := &caddy.Config{Logging: new(caddy.Logging)}
		// Hack to set a Config to caddy.Context
		// https://stackoverflow.com/a/43918797
		rs := reflect.ValueOf(&ctx).Elem()
		rf := rs.Field(2) // ctx.cfg
		ri := reflect.ValueOf(&cfg).Elem()
		rf = reflect.NewAt(rf.Type(), unsafe.Pointer(rf.UnsafeAddr())).Elem()
		rf.Set(ri)
		ri.Set(rf)

		err := prov.Provision(ctx)
		if err != nil {
			return nil, fmt.Errorf("Error during Provision: %s", err)
		}
	}

	if vali, ok := storage.(caddy.Validator); ok {
		err := vali.Validate()
		if err != nil {
			return nil, fmt.Errorf("Error during Validate: %s", err)
		}
	}

	return storage, nil
}

func ImportFiles(storage certmagic.Storage, caddyFolder string) error {
	caddyFolder, err := filepath.Abs(caddyFolder)
	if err != nil {
		return err
	}
	paths := make([]string, 0)
	err = filepath.Walk(caddyFolder, func(fullpath string, f os.FileInfo, err error) error {
		if err != nil {
			log.Printf("Error ocurred in %s: %v", fullpath, err)
		}
		if !f.IsDir() {
			paths = append(paths, fullpath)
		}
		return nil
	})

	for _, fullpath := range paths {
		path := strings.TrimPrefix(fullpath, caddyFolder)
		fmt.Printf("Importing %s...\n", path)
		binary, err := ioutil.ReadFile(fullpath)
		if err != nil {
			return err
		}
		err = storage.Store(path, binary)
		if err != nil {
			return err
		}
	}

	return err
}

func ExportFiles(storage certmagic.Storage, dest string) error {
	dest, err := filepath.Abs(dest)
	if err != nil {
		return err
	}
	keys, err := storage.List("", true)
	if err != nil {
		return err
	}
	for _, key := range keys {
		fmt.Printf("Exporting %s...\n", key)
		val, err := storage.Load(key)
		if err != nil {
			return err
		}
		path := filepath.Join(dest, key)
		err = os.MkdirAll(filepath.Dir(path), 0700)
		if err != nil {
			return err
		}
		ioutil.WriteFile(path, val, 0600)
	}
	return nil
}
