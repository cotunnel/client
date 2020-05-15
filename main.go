package main

import (
	"bufio"
	"client/app"
	"client/cog"
	"fmt"
	"github.com/kardianos/osext"
	"github.com/urfave/cli"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"
)

func main() {

	cog.SetDebug(false)

	cog.Print(cog.INFO, "COTUNNEL v"+app.Version)

	cmd := &cli.App{
		Name:    "cotunnel",
		Version: app.Version,
		Usage:   "",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "key",
				Usage:    "--key {your cotunnel client key}",
				Required: false,
			},
			&cli.BoolFlag{
				Name:     "exit",
				Usage:    "--exit true",
				Required: false,
			},
		},
		Action: func(c *cli.Context) error {
			path, err := getAppPath()
			if err != nil {
				cog.Print(cog.ERROR, "Application path not found.")

				os.Exit(0)
				return nil
			}

			options := app.Options{}
			options.Key = c.String("key")
			options.Exit = c.Bool("exit")
			options.Token = ""
			options.Path = path

			if len(options.Key) == 0 {

				tokenBytes, err := ioutil.ReadFile(options.Path + "/cotunnel.key")
				if err != nil {
					cog.Print(cog.ERROR, "cotunnel.key file not found.")
					cog.Print(cog.INFO, "Cotunnel client not logged in yet.")
					cog.Print(cog.INFO, "Visit https://www.cotunnel.com and get client key after then start the client with --key command.")
					cog.Print(cog.INFO, "If you already have Cotunnel key, please enter the key:")
					fmt.Print("> ")
					scanner := bufio.NewScanner(os.Stdin)
					if scanner.Scan() {
						key := scanner.Text()
						if len(key) == 0 {
							cog.Print(cog.ERROR, "Key not found.")
							os.Exit(0)
							return nil
						}

						options.Key = key
					}
				} else {
					options.Token = string(tokenBytes)
				}
			}

			newApp, err := app.New(&options)
			if err != nil {
				os.Exit(3)
				return nil
			}

			err = newApp.Run()
			if err != nil {
				os.Exit(4)
				return nil
			}

			registerSignals(newApp)
			return nil
		},
	}

	cmd.Run(os.Args)
}

func registerSignals(app *app.App) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(
		sigChan,
		syscall.SIGINT,
		syscall.SIGTERM,
	)

	for {
		s := <-sigChan
		switch s {
		case syscall.SIGINT, syscall.SIGTERM:
			app.Exit()
			os.Exit(5)
		}
	}
}

func getAppPath() (string, error) {
	return osext.ExecutableFolder()
}
