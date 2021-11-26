# gnatwren
The tawny-faced gnatwren (_Microbates cinereiventris_) is a very small bird
in the gnatcatcher family. This software aims to be a very small fleet
metrics and health system.

[![Image of a tawny-faced gnatwren, perched on a twig](https://github.com/firepear/gnatwren/blob/main/assets/tfgw.jpg)](https://ebird.org/species/tafgna1)  
_Image credit: Fernando Burgalin Sequeria, via ebird/the Macaulay Library_

**This software is under initial development, and is not yet suitable
for use or deployment**

## Metrics and events

Reporting of the following data is implemented for x86 and Raspberry Pi:

- CPU name and per-core frequencies
- CPU temperature (AMD K10 and Pi only)
- GPU name and temperature (AMD and NVidia)
- Total system, free, and available memory
- Uptime
- Loadavg

No events are yet being detected.

## Visualization

Not much yet, but it's in-work, along with everything else:

![Early viz](https://i.imgur.com/iIJYA4Z.png)

## Efficiency

In my most recent check across my farm, over 5.5 days of runtime the
client had used approximately 40 cpu-seconds on each node -- so a bit
under 8 cpu-seconds per day on average. Memory usage was stable at
approximately 8MB on an x86_64 system.

The aggregator has been under much more frequent development, with
large parts still be rewritten. I don't have solid statistics for it
yet.

## Installation

### Docker

To build and launch a Docker container which runs `gwgather` and an
instance of `nginx` for web monitoring, run `./build.sh`

The container has `busybox` and `sqlite` installed for diagnostics. If
needed, attach with

`docker exec -it gwgather ash`

Re-run the build script anytime. No monitoring data will be lost.


### Manual install

- `go build ./cmd/gwgather`
- `go build ./cmd/gwagent`
- `go build ./cmd/gwquery`
- Move the resulting binaries to the appropriate destinations
  - `gwgather` and `gwquery` should be in `/usr/local/bin` on the
    machine which will act as the metrics aggregator
  - `gwagent` should be in `/usr/local/bin` on each machine which will
    be sending metrics
- A systemd unit file for `gwagent` is at
  `./assets/gnatwren-agent.service`
  - It should be deployed according to systemd standards on the agent
    nodes
- A systemd unit file for `gwgather` is at
  `./assets/gnatwren-gather.service`
  - It should be deployed according to systemd standards on the
    aggregator node
- A config file for `gwagent` is at `./assets/gwagent-config.json`
  - Edit and deploy to `/etc/gnatwren/gwagent-config.json` on agent
    nodes
  - It must be readable by user `nobody`
- A config file for `gwgather` is at `./assets/gwgather-config.json`
  - Edit and deploy to `/etc/gnatwren/gwgather-config.json` on the
    aggregator node
  - It must be readable by the user `nobody`
- On the agent nodes, create the directory `/var/run/gnatwren`, which
  should be writable by `nobody`
- Make sure that the location you've defined for `gwgather`'s DB is
  writable by user `nobody`
- Enable and start the `gnatwren-gather` service on the aggregator
  node
- Enable and start the `gnatwren-agent` service on agent nodes

## Configuration

### Gather (server)

- `bind_addr`: The interface and port to bind to.Changing to
  interfaces other than `0.0.0.0` may cause failures on startup within
  Docker
- alerts
  - `late_checkin`: Seconds until a client node is considered late and
    an alert is triggered
  - `over_temp`: CPU temperature in degC which triggers an
    alert
- db
  - `location`: Path to the Gnatwren stats DB
  - `hours_retained`: How many hours of data to retain on an hourly
    (i.e. one sample per hour per node) basis
  - `days_retained`: How many days of data to retain on a daily
    (i.e. one sample per day per node) basis
- files
  - `enabled`: No current function
  - `json_location`: Path to directory where JSON stats for the web
    status page should be dumped
  - `json_interval`: Frequency, in seconds, of JSON dumps

### Client

- `gather_addr`: IP addr and port where the gather daemon is listening
- `active`: No current function
- `intervals`: The set of intervals, in seconds, which will be
  selected from after each report is made. The default set is the
  primes around 30 and 45, resulting in (on average) 1.69 updates per
  minute, while minimizing simultaneous updates
