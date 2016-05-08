package main

import (
	"bufio"
	"crypto/tls"
	"database/sql"
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jfmyers9/gotta-track-em-all/db"
	"github.com/jfmyers9/gotta-track-em-all/handlers"
	"github.com/jfmyers9/gotta-track-em-all/models"
	"github.com/jfmyers9/gotta-track-em-all/watcher"
	"github.com/pivotal-golang/lager"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/grouper"
	"github.com/tedsuo/ifrit/http_server"
	"github.com/tedsuo/ifrit/sigmon"

	_ "github.com/lib/pq"
)

var pokemonCSV = flag.String(
	"pokemonCSV",
	"",
	"path to pokemon csv",
)

var listenAddress = flag.String(
	"listenAddress",
	"",
	"Address to listen for requests on",
)

var dbConnectionString = flag.String(
	"dbConnectionString",
	"",
	"The connection string to the postgres db",
)

func main() {
	flag.Parse()
	logger := lager.NewLogger("gotta-track-em-all")

	rand.Seed(time.Now().UnixNano())

	sink := lager.NewReconfigurableSink(lager.NewWriterSink(os.Stdout, lager.DEBUG), lager.DEBUG)
	logger.RegisterSink(sink)

	pokemonList, err := parsePokemonCSV(*pokemonCSV)
	if err != nil {
		logger.Error("failed-to-parse-pokemon", err)
		os.Exit(1)
	}

	sqlConn, err := sql.Open("postgres", *dbConnectionString)
	if err != nil {
		logger.Error("failed-to-construct-sql-conn", err)
		os.Exit(1)
	}

	err = sqlConn.Ping()
	if err != nil {
		logger.Error("failed-to-connect-to-database", err)
		os.Exit(1)
	}

	d := db.NewDB(sqlConn)

	err = d.RunMigrations(logger)
	if err != nil {
		logger.Error("failed-to-run-migrations", err)
		os.Exit(1)
	}

	handler, err := handlers.NewHandler(logger, d)
	if err != nil {
		logger.Error("failed-to-construct-handlers", err)
		os.Exit(1)
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	httpClient := &http.Client{Transport: tr}

	members := grouper.Members{
		{"api", http_server.New(*listenAddress, handler)},
		{"watcher", watcher.NewWatcher(logger, d, httpClient, pokemonList)},
	}

	group := grouper.NewOrdered(os.Interrupt, members)

	monitor := ifrit.Invoke(sigmon.New(group))

	logger.Info("started")

	err = <-monitor.Wait()
	if err != nil {
		logger.Error("exited-with-failure", err)
		os.Exit(1)
	}

	logger.Info("exited")
}

func parsePokemonCSV(path string) ([]*models.PokemonEntry, error) {
	var cumWeight float64 = 0
	pokemonList := []*models.PokemonEntry{}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		row := strings.Split(string(scanner.Text()), ",")
		if len(row) != 3 {
			println("what?")
			continue
		}

		index, err := strconv.Atoi(row[0])
		if err != nil {
			fmt.Println("messed up")
			continue
		}

		weight, err := strconv.ParseFloat(row[2], 64)
		if err != nil {
			fmt.Println("messed up more")
			continue
		}

		cumWeight += weight

		entry := models.PokemonEntry{
			Index:  index,
			Name:   strings.Title(row[1]),
			Weight: cumWeight,
		}
		pokemonList = append(pokemonList, &entry)
	}

	return pokemonList, scanner.Err()
}
