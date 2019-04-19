package main

import (
	"log"
	"os"

	"fmt"
	"path"
	"time"

	"github.com/dfinity/go-dfinity-crypto/bls"
	"github.com/keep-network/keep-core/cmd"
	"github.com/urfave/cli"
)

const defaultConfigPath = "./config.toml"
const defaultConsulPath = ""

var (
	version  string
	revision string

	configPath string
	consulPath string
)

func main() {
	if version == "" {
		version = "unknown"
	}
	if revision == "" {
		revision = "unknown"
	}

	if err := bls.Init(bls.CurveSNARK1); err != nil {
		log.Fatal("Failed to initialize BLS.", err)
	}

	app := cli.NewApp()
	app.Name = path.Base(os.Args[0])
	app.Usage = "CLI for The Keep Network"
	app.Description = "Command line interface (CLI) for running a Keep provider"
	app.Compiled = time.Now()
	app.Authors = []cli.Author{
		{
			Name:  "Keep Network",
			Email: "info@keep.network",
		},
	}
	app.Version = fmt.Sprintf("%s (revision %s)", version, revision)
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "config,c",
			Value:       defaultConfigPath,
			Destination: &configPath,
			Usage:       "full path to the configuration file",
		},
		cli.StringFlag{
			Name:        "consul,u",
			Value:       defaultConsulPath,
			Destination: &consulPath,
			Usage:       "<ConsulServer>:<Port>",
		},
	}
	app.Commands = []cli.Command{
		cmd.SmokeTestCommand,
		cmd.StartCommand,
		cmd.RelayCommand,
		cmd.PingCommand,
		{
			Name:  "print-info",
			Usage: "Prints keep client information",
			Action: func(c *cli.Context) error {
				printInfo(c)
				return nil
			},
		},
	}

	cli.AppHelpTemplate = fmt.Sprintf(`%s
ENVIRONMENT VARIABLES:
   KEEP_ETHEREUM_ACCOUNT_KEYFILEPASSWORD    keep client password

`, cli.AppHelpTemplate)

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func printInfo(c *cli.Context) {
	fmt.Printf("Keep client: %s\n\n"+
		"Description: %s\n"+
		"version:     %s\n"+
		"revision:    %s\n"+
		"Config Path: %s\n",
		c.App.Name,
		c.App.Description,
		version,
		revision,
		c.GlobalString("config"),
	)
}
