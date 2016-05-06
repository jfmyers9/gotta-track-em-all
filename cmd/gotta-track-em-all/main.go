package main

import (
	"crypto/tls"
	"database/sql"
	"flag"
	"net/http"
	"os"

	"github.com/jfmyers9/gotta-track-em-all/db"
	"github.com/jfmyers9/gotta-track-em-all/handlers"
	"github.com/jfmyers9/gotta-track-em-all/watcher"
	"github.com/pivotal-golang/lager"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/grouper"
	"github.com/tedsuo/ifrit/http_server"
	"github.com/tedsuo/ifrit/sigmon"

	_ "github.com/lib/pq"
)

var trackerAPIToken = flag.String(
	"trackerAPIToken",
	"",
	"API Token used to access the tracker api",
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

	sink := lager.NewReconfigurableSink(lager.NewWriterSink(os.Stdout, lager.DEBUG), lager.DEBUG)
	logger.RegisterSink(sink)

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

	err = d.CreateInitialSchema(logger)
	if err != nil {
		logger.Error("failed-creating-schema", err)
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
		{"watcher", watcher.NewWatcher(logger, d, httpClient)},
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
