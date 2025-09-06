# gnatwren
The tawny-faced gnatwren (_Microbates cinereiventris_) is a very small bird
in the gnatcatcher family. This software aims to be a very small fleet
metrics and health system.

[![Image of a tawny-faced gnatwren, perched on a twig](https://github.com/firepear/gnatwren/blob/main/assets/tfgw.jpg)](https://ebird.org/species/tafgna1)  
_Image credit: Fernando Burgalin Sequeria, via ebird/the Macaulay Library_

**This software is under initial development, and is not yet suitable
for use or deployment**

## Metrics and events

Reporting of the following data is implemented:

- CPU name and average frequencies
- CPU temperature (AMD K10, Intel, and Pi)
- Uptime
- Loadavg
- Total, free, and available memory
- GPU name, temperature, fan speed, and power stats (Nvidia and AMD; name for Intel)
- Time since last check-in

The web status page also displays:

- CPU temperature
  - Warning status if over configured value
  - Critical status if over configured value
- Last check-in
  - Critical status if more than 210s overdue
  - Warning if a node ceases to be present in current data

![wip viz](https://i.imgur.com/fWPAxVU.png)



# Configuration

First, do `cp assets/*.json .`

This will create copies of the stock `gwgather` and `gwagent` config
files which you can edit and use going forward.

## gwgather

- `bind_addr`: The interface and port to bind to. Changing to
  values other than `0.0.0.0` may cause failures on container startup
- db
  - `location`: Path to the Gnatwren stats DB
  - `hours_retained`: How many hours of data to retain on an hourly
    (i.e. one sample per hour per node) basis
  - `days_retained`: How many days of data to retain on a daily
    (i.e. one sample per day per node) basis
- files
  - `json_location`: Path to directory where JSON stats for the web
    status page should be dumped
  - `json_interval`: Frequency, in seconds, of JSON dumps
- log
  - `file`: The file to write output to
  - `level`: Logging level ("fatal", "error", "conn", "debug")
- ui (Web UI config)
  - `title`: Title for stats page
  - `temp_hi_cpu`: Temp at which to display the high warning for CPU
  - `temp_crit_cpu`: Temp at which to display the critical warning for CPU

## gwagent

- `gather_addr`: IP addr and port where the gather daemon is listening
- `active`: No current function
- `intervals`: The set of intervals, in seconds, which will be
  selected from after each report is made. The default set is the
  primes between 30 and 50, resulting in (on average) 1.51 updates per
  minute, while minimizing simultaneous updates
- `gpu`: Specify which GPU to monitor, for non-Nvidia GPUs. Defaults
  to empty string, which searches for the first defined GPU in the
  system. Example: `card2` (as in `/sys/class/drm/card2`)
- `workdir`: Directory where updates are stashed in the event of
  network issues



# Installation

## gwgather via docker/podman

To build and launch a container which runs `gwgather` and an instance
of `nginx` for web monitoring, run `./build.sh` (via `sudo` if you're
using `podman` without rootless containers). The build script will use
the copy of the `gwgather` config file that you've edited, or it will
create this copy if you haven't already.

Re-run the build script anytime. No monitoring data will be lost.

The container has `busybox` and `sqlite` installed for diagnostics. If
needed, attach with

`docker exec -it gwgather ash`

## Manual install

### gwgather

Examining the build script and Dockerfile will show everything that is
needed and how it's done. It's very straightforward and should be
adaptable to any situation without much effort.

### gwagent

- `go build ./cmd/gwagent`
- `mv ./gwagent /usr/local/bin`
  - Repeat this on any systems to be monitored
- A systemd unit file for `gwagent` is at
  `./assets/gnatwren-agent.service` and should be copied to
  `/etc/systemd/system` on any systems to be monitored
- Edit `agent-config.json` and then deploy it to `/etc/gnatwren/` on
  all nodes to be monitoried
  - It must be readable by user `nobody`
- On all agent nodes, create the directory `/var/run/gnatwren`, which
  should be writable by `nobody`
- Enable and start the `gnatwren-agent` service on agent nodes
  - You may need to do `sudo systemctl daemon-reload` first
