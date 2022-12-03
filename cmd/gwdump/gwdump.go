package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/firepear/gnatwren/internal/data"
	_ "github.com/mattn/go-sqlite3"
)


var (
	// gwgather config
	config data.GatherConfig
	// export periods
	periods = [3]string{"current", "hourly", "daily"}
	// db handle
	db  *sql.DB
)

func exportOverview() error {
	machStatus, err := dbGetOverview()
	if err != nil {
		return err
	}
	var filename string
	var data []byte

	filename = fmt.Sprintf("%s/overview.json", config.Files.JsonLoc)
	data, _ = json.Marshal(*machStatus)
	err = os.WriteFile(filename, data, 0644)
	return err
}

func exportCpuTemps(duration string) error {
	cpuTemps, err := dbGetCPUStats(duration)
	if err != nil {
		return err
	}

	var filename string
	filename = fmt.Sprintf("%s/cputemps-%s.json", config.Files.JsonLoc, duration)
	cpuTempsj, _ := json.Marshal(*cpuTemps)
	err = os.WriteFile(filename, cpuTempsj, 0644)
	return err
}

func exportGpuTemps(duration string) error {
	gpuTemps, err := dbGetGPUStats(duration)
	if err != nil {
		return err
	}

	var filename string
	filename = fmt.Sprintf("%s/gputemps-%s.json", config.Files.JsonLoc, duration)
	gpuTempsj, _ := json.Marshal(*gpuTemps)
	err = os.WriteFile(filename, gpuTempsj, 0644)
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

	// run overview
	err = exportOverview()
	if err != nil {
		log.Printf("couldn't export overview to json: %s\n", err)
	}
	// export CPU temps
	for _, period := range periods {
		err = exportCpuTemps(period)
		if err != nil {
			log.Printf("couldn't export %s cpu temps to json: %s\n",
				period, err)
		}
		time.Sleep(1)
		err = exportGpuTemps(period)
		if err != nil {
			log.Printf("couldn't export %s gpu temps to json: %s\n",
				period, err)
		}
		time.Sleep(1)
	}
}
