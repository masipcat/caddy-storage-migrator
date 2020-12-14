package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/caddyserver/certmagic"
	migrator "github.com/masipcat/caddy-storage-migrator"
)

type command struct {
	Usage string
	NArgs int
	Func  func(args []string)
}

func main() {
	configFile := flag.String("config", "", "Path to JSON file")
	flag.Parse()
	config, err := loadConfig(*configFile)
	if err != nil {
		fmt.Printf("%v", err)
		os.Exit(1)
	}

	commands := map[string]*command{
		"import": {
			Usage: "migrator [-config FILE] import SOURCE STORAGE",
			NArgs: 2,
			Func: func(args []string) {
				source, name := args[0], args[1]
				s, err := migrator.InitStorage(name, config)
				if err != nil {
					fmt.Printf("Failed to init storage: %v", err)
					os.Exit(1)
				}
				migrator.ImportFiles(s.(certmagic.Storage), source)
			},
		},
		"export": {
			Usage: "migrator [-config FILE] export STORAGE DEST",
			NArgs: 2,
			Func: func(args []string) {
				name, dest := args[0], args[1]
				s, err := migrator.InitStorage(name, config)
				if err != nil {
					fmt.Printf("Failed to init storage: %v", err)
					os.Exit(1)
				}
				migrator.ExportFiles(s.(certmagic.Storage), dest)
			},
		},
	}

	if flag.NArg() > 0 {
		cmdName := flag.Arg(0)
		if cmd, found := commands[cmdName]; found {
			cmd.Run(flag.Args()[1:])
			return
		}
	}

	fmt.Println("Usage:")
	for _, c := range commands {
		fmt.Printf("\t\t%s\n", c.Usage)
	}
}

func (c *command) Run(args []string) {
	minArgs := len(args)
	if minArgs >= c.NArgs {
		c.Func(args)
		os.Exit(0)
	} else {
		fmt.Printf("Missing %d argument(s). Usage:\n\n\t\t%s\n\n", c.NArgs-minArgs, c.Usage)
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
		return nil, fmt.Errorf("Key 'storage' not found in json")
	}
	return data, nil
}
