package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/caddyserver/certmagic"
	migrator "github.com/masipcat/caddy-storage-migrator"
)

func main() {
	var configFile string
	flag.StringVar(&configFile, "config", "", "Path to json file")

	importCmd := flag.NewFlagSet("import", flag.ExitOnError)
	importCmd.Usage = func() {
		fmt.Println("migrator [-config file.json] import SOURCE STORAGE_NAME")
	}

	exportCmd := flag.NewFlagSet("export", flag.ExitOnError)
	exportCmd.Usage = func() {
		fmt.Println("migrator [-config file.json] export STORAGE_NAME DEST")
	}

	flag.Parse()

	config, err := loadConfig(configFile)
	if err != nil {
		log.Fatalf("%v", err)
	}

	var cmd = ""
	if flag.NArg() > 0 {
		cmd = flag.Arg(0)
	}

	switch cmd {
	case "import":
		importCmd.Parse(flag.Args()[1:])
		if importCmd.NArg() < 2 {
			importCmd.Usage()
			os.Exit(1)
		}
		s, err := migrator.InitStorage(importCmd.Arg(1), config)
		if err != nil {
			log.Fatalf("Failed to init storage: %v", err)
		}
		migrator.ImportFiles(s.(certmagic.Storage), importCmd.Arg(0))
	case "export":
		exportCmd.Parse(flag.Args()[1:])
		if exportCmd.NArg() < 2 {
			exportCmd.Usage()
			os.Exit(1)
		}
		s, err := migrator.InitStorage(exportCmd.Arg(0), config)
		if err != nil {
			log.Fatalf("Failed to init storage: %v", err)
		}
		migrator.ExportFiles(s.(certmagic.Storage), exportCmd.Arg(1))
	default:
		importCmd.Usage()
		exportCmd.Usage()
		os.Exit(1)
	}
}

func loadConfig(configFile string) ([]byte, error) {
	if configFile == "" {
		return []byte{}, nil
	}
	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("Error reading config: %v", err)
	}
	var config map[string]json.RawMessage
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("Couldn't unmarshal json: %v", err)
	}
	data, found := config["storage"]
	if !found {
		return nil, fmt.Errorf("key 'storage' not found")
	}
	return data, nil
}
