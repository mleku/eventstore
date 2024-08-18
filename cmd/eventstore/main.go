package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/fiatjaf/cli/v3"
	"github.com/mleku/eventstore"
	"github.com/mleku/eventstore/badger"
)

var db eventstore.Store

var app = &cli.Command{
	Name:      "eventstore",
	Usage:     "a CLI for all the eventstore backends",
	UsageText: "eventstore -d ./data/sqlite <query|save|delete> ...",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "store",
			Aliases:  []string{"d"},
			Usage:    "path to the database file or directory or database connection uri",
			Required: true,
		},
		&cli.StringFlag{
			Name:    "type",
			Aliases: []string{"t"},
			Usage:   "store type ('sqlite', 'lmdb', 'bolt', 'badger', 'postgres', 'mysql', 'elasticsearch')",
		},
	},
	Before: func(ctx context.Context, c *cli.Command) error {
		path := c.String("store")
		typ := c.String("type")
		if typ != "" {
			// bypass automatic detection
			// this also works for creating disk databases from scratch
		} else {
			// try to detect based on url scheme
			switch {
			case strings.HasPrefix(path, "postgres://"), strings.HasPrefix(path,
				"postgresql://"):
				typ = "postgres"
			case strings.HasPrefix(path, "mysql://"):
				typ = "mysql"
			case strings.HasPrefix(path, "https://"):
				// if we ever add something else that uses URLs we'll have to modify this
				typ = "elasticsearch"
			case strings.HasSuffix(path, ".conf"):
				typ = "strfry"
			default:
				// try to detect based on the form and names of disk files
				dbname, err := detect(path)
				if err != nil {
					if os.IsNotExist(err) {
						return fmt.Errorf(
							"'%s' does not exist, to create a store there specify the --type argument",
							path)
					}
					return fmt.Errorf("failed to detect store type: %w", err)
				}
				typ = dbname
			}
		}

		switch typ {
		case "badger":
			db = &badger.BadgerBackend{Path: path, MaxLimit: 1_000_000}
		case "":
			return fmt.Errorf("couldn't determine store type, you can use --type to specify it manually")
		default:
			return fmt.Errorf("'%s' store type is not supported by this CLI", typ)
		}

		return db.Init()
	},
	Commands: []*cli.Command{
		queryOrSave,
		query,
		save,
		delete_,
	},
	DefaultCommand: "query-or-save",
}

func main() {
	if err := app.Run(context.Background(), os.Args); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
