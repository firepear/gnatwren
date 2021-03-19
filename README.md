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

- CPU name, temperature (AMD K10 and Pi only), and per-core frequencies
- Total and available memory
- Uptime
- Loadavg

No events are yet being detected.

## Installation

My project [Homefarm](https://github.com/firepear/homefarm), has a
pair of Ansible playbooks which could be used as a base for your own
customized install if you're using Ansible or a similar config
management tool:

- [Control node playbook](https://github.com/firepear/homefarm/blob/master/gnatwren-control.yml) (builds gwgather and gwquery)
- [Compute node playbook](https://github.com/firepear/homefarm/blob/master/gnatwren-nodes.yml) (builds gwagent and deploys it to nodes)

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
- A config file for `gwagent` is at `./assets/agent-config.json`
  - Edit and deploy to `/etc/gnatwren/gwagent-config.json` on agent
    nodes
  - It must be readable by user `nobody`
- A config file for `gwgather` is at `./assets/gather-config.json`
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
