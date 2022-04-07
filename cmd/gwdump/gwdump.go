package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"log"
	"os"
	//	"os/signal"
	//"strconv"
	"strings"
	"sync"
	//"syscall"
	//"time"

	//"github.com/firepear/petrel"
	"github.com/firepear/gnatwren/internal/data"
	_ "github.com/mattn/go-sqlite3"
)


var (
	// gwgather config
	config data.GatherConfig
	// db handle
	db *sql.DB
	// nodeStatus holds the last check-in time of nodes running
	// agents. mux is its lock
	nodeStatus = map[string][2]int64{}
	mux sync.RWMutex
)

func exportOverview() error {
	machStats, err := dbGetOverview()
	if err != nil {
		return err
	}
	var filename strings.Builder
	var data []byte
	filename.WriteString(config.Files.JsonLoc)
	filename.WriteString("/overview.json")
	//data = append(data, []byte("{")...)
	//i := 1
	//for host, metrics := range *machStats {
	//	data = append(data, []byte(`"`)...)
	//	data = append(data, []byte(host)...)
	//	data = append(data, []byte(`": {"TS":`)...)
	//	data = append(data, []byte(strconv.FormatInt(metrics.TS, 10))...)
	//	data = append(data, []byte(`,"Payload":`)...)
	//	data = append(data, []byte(metrics.Payload)...)
	//	data = append(data, []byte(`}`)...)
	//	if i < len(*machStats) {
	//		data = append(data, []byte(`,`)...)
	//	}
	//	i++
	//}
	//data = append(data, []byte("}")...)
	data, _ = json.Marshal(*machStats)
	err = os.WriteFile(filename.String(), data, 0644)
	return err
}

func exportCpuTemps(duration string) error {
	cpuTemps, err := dbGetCPUStats(duration)
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
	dbLoadNodeStatus()
	defer db.Close()

	err = exportOverview()
	if err != nil {
		log.Printf("couldn't export to json: %s\n", err)
	}
	err = exportCpuTemps("current")
	if err != nil {
		log.Printf("couldn't export to json: %s\n", err)
	}
}
