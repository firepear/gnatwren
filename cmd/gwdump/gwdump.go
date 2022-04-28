package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/firepear/gnatwren/internal/data"
	_ "github.com/mattn/go-sqlite3"
)


var (
	// gwgather config
	config data.GatherConfig
	// db handle
	db *sql.DB
	mux sync.RWMutex
)

func exportOverview() error {
	machStatus, err := dbGetOverview()
	if err != nil {
		return err
	}
	var filename strings.Builder
	var data []byte

	filename.WriteString(config.Files.JsonLoc)
	filename.WriteString("/overview.json")
	data, _ = json.Marshal(*machStatus)
	err = os.WriteFile(filename.String(), data, 0644)
	return err
}

func exportCpuTemps(duration string) error {
	cpuTemps, err := dbGetCPUStats(duration)
	//log.Printf("cputemps: %v\n", cpuTemps)
	if err != nil {
		return err
	}

	var filename strings.Builder
	filename.WriteString(config.Files.JsonLoc)
	filename.WriteString("/cputemps-")
	filename.WriteString(duration)
	filename.WriteString(".json")
	cpuTempsj, _ := json.Marshal(*cpuTemps)
	err = os.WriteFile(filename.String(), cpuTempsj, 0644)
	return err
}

func main() {
	// find out where the gwagent config file is and read it in
	var configfile string
	flag.StringVar(&configfile, "config", "/etc/gnatwren/gather.json", "Location of the gwgather config file")
	flag.Parse()

	configstr, err := os.ReadFile(configfile)
	if err != nil {
		log.Fatal(err)
	}
	err = json.Unmarshal(configstr, &config)
	if err != nil {
		log.Fatal(err)
	}

	// initialize database
	db, err = dbSetup(config.DB.Loc)
	if err != nil {
		log.Fatalf("sqlite: can't init db: %s", err)
	}
	defer db.Close()

	err = exportOverview()
	if err != nil {
		log.Printf("couldn't export overview to json: %s\n", err)
	}
	err = exportCpuTemps("current")
	if err != nil {
		log.Printf("couldn't export current cpu temps to json: %s\n", err)
	}
}
