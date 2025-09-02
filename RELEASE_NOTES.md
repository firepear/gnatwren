## 0.20.0 (2025-09-xx)

- New config points for gather and webui (see README)
- Fixes for AMD GPUs
- Added `-once` flag to gwagent, which is useful for testing (gathers
  metrics, dumps to stdout, and exits)
- Updated for Petrel v0.40
- Integrated logging with Petrel (mostly)
- Updated echarts version for web UI

## 0.19.1 (2023-07-23)

- Changes for Petrel v0.36
- Changes for Nvidia GPU data collection (because we have to scrape
  `nvidia-smi` rather than use `sysfs`)


## 0.19.0 (2022-12-18)

- Petrel dependency updated to 0.36.0

### gwgather

- CPU and GPU temps now exported to current, hourly, and daily files
- `jq` added to container image

### gwagent

- Nvidia cards too old for the installed driver now have their model
  name collected from `lspci`

### web

- Host missing time calculation fixed
- Javascript split into its own source file



## 0.18.0 (2022-12-03)

### gwgather

- DB table `daily` was accumulating values every 48h rather than
  every 24. This has been fixed
- DB retention now honors config settings



## 0.17.3 (2022-07-16)

### web

- Removed `localhost` and port number from URL construction inside
  script; everything is now purely relative to the httpd root



## 0.17.2 - 2022-06-11

### agent

- Metrics stash moved from /var/run to /var/lib



## 0.17.1 - 2022-05-12

### web

- Fixes bad dates in missing host notification
- Give page a title, finally


## 0.17.0 - 2022-05-08

### gwagent

- Improved Intel CPU name handling
- Initial pass at info gathering for Intel iGPUs

### web

- Further shortening of CPU descs
- Re-added CPU thread count
- Placeholders for null values from GPUs

### all

- `gofmt` fixes



## 0.16.0 - 2022-05-07

### gwgather

- nodeStatus now begins with all hosts from all db tables
- A crashing bug due to new nodes coming online was fixed
- The `alerts` section of the config file has been removed, along with
  its struct. A new section, `log` has been added

- gwgather now logs to a file, and log level is settable via
      config file

### gwagent

- CPU temps are now pulled from Intel processors



## 0.15.0 - 2022-04-29

### web

- Alerts are now set for nodes which have been seen, but have fallen
  out of the current system data


## 0.14.0 - 2022-04-28

### gwgather

- Simplified data handling for adding metrics to DB
- Dead code removal

### gwdump

- Removed unneeded node update tracking code + vars

### web

- Hovering over 'Last' now shows update time as a tooltip



## 0.13.0 - 2022-04-24

### gwagent

- Enabled GPU data gathering for AMD GPUs
- Tweaked data formating for Nvidia GPUs; all GPUs will be have
  identical formatting from here on out



## 0.12.0 - 2022-04-23

### gwagent

- Enabled GPU info reporting, beginning with Nvidia cards
- Metrics stashing dir is (re)created if it doesn't exist, and there
  is an attept to stash

### web

- GPU data now shown in system stats



## 0.11.0 - 2022-04-08

### gwagent

- Removed average clock speed from CPUdata struct

### gwgather

- Config file attribute `over_temp` changed to `temp_warn`, and new
  attribute `temp_crit` was added

### gwdump

- Changed JSON output filenames
- Generally simplified data handling

### web

- Each node's time since last checkin is now shown
- Hovering over Clock displays a tooltip with per-core clocks
- CPU temp displays a warning status at 80C; critical at 90C
- Last seen displays a critical status if a node's checkin time is
  more than 210 seconds ago (90s plus the data refresh period of 120s)



## 0.10.0 - 2021-11-27

- Finished removal of gwquery and its supporting code
- Removed JSON dump functionality from gwgather; it has been placed in
  a new utility named gwdump which runs periodicaly in gwgather's
  container



## 0.9.1 - 2021-11-27

- Average clock now computed and stored in DB
- Loadavg is now a single value (1 minute)
- Improvements in avoiding unnecessary JSON (un)marshalling
  stemming from hasty BadgerDB -> SQLite rewrite
- Go sqlite lib now built with JSON1 extension



## 0.9.0 - 2021-11-26

- Gnatwren is now fully discrete from Homefarm
- gwgather is containerized
- Small fixes to CPU temp reporting for k10 driver



## 0.8.1 - 2021-04-26

- Machine archtecture is now gathered and reported
- Hostname is only checked once



## 0.8.0 - 2021-03-22

- gwgather now exports CPU temp data to a JSON file every 5
  minutes
- An HTML page which renders CPU temp data using the echarts lib
  has been added -- the beginning of data visualization efforts
- 'dbstatus' added to gwquery



## 0.7.0 - 2021-03-21

- gwgather now makes DB GC calls on startup, and every 45 minutes
  subsequently
- Petrel handlers split out into petrel.go



## 0.6.1 - 2021-03-19

- gwgather's nodeStatus map now stores two timestamps: most recent
  checkin and newest metrics. When nodes are reporting stowed metrics,
  these will differ

- gwquery updated to handle AgentStatus, a new type derived from
  AgentPayload, which includes the most recent checkin timestamp



## 0.6.0 - 2021-03-19

- gwgather now stores the past 24h of reported metrics
- Crash-on-halt fixed in gwgather



## 0.5.0 - 2021-03-15

- gwagent now stows unsendable metrics to a file, instead of
  dying. Stowed metrics are sent when connectivity to gwgather
  recovers
- Go 1.16 is now required, as ioutil dependencies have been removed



## 0.4.0 - 2021-03-07

- gwgather now exists, and aggregates/reports data
- gwagent is now properly an agent (client reporting to a server),
  rather than a server
- gwquery now requests data from gwgather rather than polling all
  agents



## 0.3.0 - 2021-01-11

- gwagent now takes a config file



## 0.2.1 - 2021-01-04

- Changed memory to show available and freeable



## 0.2.0 - 2021-01-03

- Added loadavg



## 0.1.0 - 2021-01-02

- Pathfinder/demo implmentation; production design will work in a
  completely different way
- Agent: memory, uptime, cpu clocks and temp gathering on AMD Ryzen
  and Raspberry Pi. Uses only /proc and /sys data; no calls to
  external processes; no third-party golang deps
- Query: simple, static, "on my machine" only implementation. Can
  gather metrics from all machines in 0.14 seconds

