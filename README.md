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

The following events are shown on the web status page:

- CPU temperature
  - Warning status if CPU over 80C
  - Critical status if CPU over 90C
- Last check-in
  - Critical status if more than 210s overdue
  - Warning if a node ceases to be present in current data

## Visualization

This screenshot shows some new-since v0.11.0 features: per-core clocks
when hovering over the averaged clock, aa critical warning indicator
for machines that have not checked in recently, and GPU data for
Nvidia and AMD cards.

![wip viz](https://i.imgur.com/fWPAxVU.png)

## Efficiency

In my most recent check across my farm, over 5.5 days of runtime the
client had used approximately 40 cpu-seconds on each node -- so a bit
under 8 cpu-seconds per day on average. Memory usage was stable at
approximately 8MB on an x86_64 system.

The aggregator has been under much more frequent development, with
large parts still be rewritten. I don't have solid statistics for it
yet.

## Installation

### gwgather via Docker

To build and launch a Docker container which runs `gwgather` and an
instance of `nginx` for web monitoring, run `./build.sh`

Re-run the build script anytime. No monitoring data will be lost.

The container has `busybox` and `sqlite` installed for diagnostics. If
needed, attach with

`docker exec -it gwgather ash`

### gwagent via Ansible

My Homefarm project contains an Ansible playbook which will build
gwagent and deploy it to a set of nodes.

### Manual install

#### gwgather

Examining the build script and Dockerfile will show everything that is
needed and how it's done. It's very straightforward and should be
adaptable to any situation without much effort.

#### gwagent

- `go build ./cmd/gwagent`
- `mv ./gwagent /usr/local/bin`
- A systemd unit file for `gwagent` is at
  `./assets/gnatwren-agent.service`
  - It should be deployed according to systemd standards on the agent
    nodes
- A config file for `gwagent` is at `./assets/gwagent-config.json`
  - Edit and deploy to `/etc/gnatwren/agent-config.json` on agent
    nodes
  - It must be readable by user `nobody`
- On the agent nodes, create the directory `/var/run/gnatwren`, which
  should be writable by `nobody`
- Enable and start the `gnatwren-agent` service on agent nodes

## Configuration

### gwgather

- `bind_addr`: The interface and port to bind to.Changing to
  interfaces other than `0.0.0.0` may cause failures on startup within
  Docker
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

### gwagent

- `gather_addr`: IP addr and port where the gather daemon is listening
- `active`: No current function
- `intervals`: The set of intervals, in seconds, which will be
  selected from after each report is made. The default set is the
  primes between 30 and 50, resulting in (on average) 1.51 updates per
  minute, while minimizing simultaneous updates
