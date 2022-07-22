package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type printlner func(v ...any)

func main() {
	var (
		flagHelp    = flag.Bool("help", false, "shows usage")
		flagListDBs = flag.Bool("list-dbs", false, "List all available DBs (used for auto-completion)")
	)
	flag.BoolVar(flagHelp, "h", false, "shows usage")
	flag.Parse()
	if *flagHelp {
		usage("")
	}
	if *flagListDBs { // for auto-completion
		for dbName := range mustReadDatabasesConfigFile() {
			fmt.Print(dbName, " ")
		}
		fmt.Println()
		os.Exit(0)
	}

	databases := mustReadDatabasesConfigFile()

	if len(os.Args[1:]) == 0 {
		usage("Target database unspecified; where should I run the query?")
	}

	var query string
	var databasesArgs []string

	stat, err := os.Stdin.Stat()
	if err != nil {
		log.Fatalf("Couldn't os.Stdin.Stat(): %v", err)
	}
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		// Stdin is a terminal. The last argument is the SQL.
		if len(os.Args) < 3 {
			usage("No SQL to run. Exiting.")
		}
		query = os.Args[len(os.Args)-1]
		databasesArgs = os.Args[1 : len(os.Args)-1]
	} else {
		query = readQuery(os.Stdin)
		databasesArgs = os.Args[1:]
	}

	if len(query) <= 3 {
		usage("No SQL to run. Exiting.")
	}

	os.Exit(_main(databases, databasesArgs, query, log.New(os.Stdout, "", 0).Println))
}

func _main(databases map[string]database, databasesArgs []string, query string, println printlner) int {
	targetDatabases := []string{}
	for _, k := range databasesArgs {
		if _, ok := databases[k]; k != "all" && !ok {
			usage("Target database unknown: [%v]", k)
		}
		if k == "all" {
			targetDatabases = nil
			for k := range databases {
				targetDatabases = append(targetDatabases, k)
			}
			break
		}
		targetDatabases = append(targetDatabases, k)
	}

	quitContext, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(len(targetDatabases))

	sqlRunner := mustNewSQLRunner(quitContext, println, query, len(targetDatabases) > 1)

	returnCode := 0
	for _, k := range targetDatabases {
		go func(db database, k string) {
			defer wg.Done()
			if r := sqlRunner.runSQL(db, k); !r {
				returnCode = 1
			}
		}(databases[k], k)
	}

	wg.Wait()
	return returnCode
}
