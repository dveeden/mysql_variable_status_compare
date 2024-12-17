package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"regexp"
	"slices"

	_ "github.com/go-sql-driver/mysql"
)

func getGlobalVarNames(ctx context.Context, db *sql.DB) ([]string, string) {
	rows, err := db.QueryContext(ctx, "SHOW GLOBAL VARIABLES")
	if err != nil {
		panic(err)
	}

	names := make([]string, 0)
	version := ""
	var name, value string
	for rows.Next() {
		rows.Scan(&name, &value)
		names = append(names, name)
		if name == "version" {
			version = value
		}
	}

	return names, version
}

func main() {
	var srvA = flag.String("server-a", "", "DSN for server A")
	var srvB = flag.String("server-b", "", "DSN for server B")
	var excl = flag.String("exclude", "^(tidb|tikv|tiflash|mysqlx|ndb|ndbinfo)_", "Regex for variables to exclude")
	var incl = flag.String("include", ".*", "Regex for variables to include")
	flag.Parse()

	if *srvA == "" || *srvB == "" {
		flag.Usage()
		os.Exit(1)
	}

	exclude := regexp.MustCompile(*excl)
	include := regexp.MustCompile(*incl)

	dbA, err := sql.Open("mysql", *srvA)
	if err != nil {
		panic(fmt.Errorf("Failed to connect to server A: %v", err))
	}

	dbB, err := sql.Open("mysql", *srvB)
	if err != nil {
		panic(fmt.Errorf("Failed to connect to server B: %v", err))
	}

	ctx := context.Background()
	varsA, verA := getGlobalVarNames(ctx, dbA)
	varsB, verB := getGlobalVarNames(ctx, dbB)

	fmt.Printf("Server A: %v\nServer B: %v\n\n", verA, verB)

	for _, v := range varsA {
		if exclude.MatchString(v) {
			continue
		}
		if !include.MatchString(v) {
			continue
		}
		if !slices.Contains(varsB, v) {
			fmt.Printf("only on A: %s\n", v)
		}
	}

	fmt.Printf("\n")

	for _, v := range varsB {
		if exclude.MatchString(v) {
			continue
		}
		if !include.MatchString(v) {
			continue
		}
		if !slices.Contains(varsA, v) {
			fmt.Printf("only on B: %s\n", v)
		}
	}
}
