# LoadMark

**Kubernetes performance benchmarking with k6 and Prometheus.**

LoadMark is a lightweight performance benchmarking tool that runs staged [k6](https://k6.io/) load tests and collects Kubernetes resource metrics from [Prometheus](https://prometheus.io/) during each stage.

Instead of looking only at request latency and throughput, LoadMark helps correlate increasing application load with what is happening inside the infrastructure.

For example:

```text
10 RPS
  ↓
25 RPS
  ↓
50 RPS
  ↓
100 RPS
```

LoadMark can collect resource metrics for each stage:

```text
                10 RPS    25 RPS    50 RPS    100 RPS
Service 1 CPU      ...       ...       ...        ...
Service 2 CPU      ...       ...       ...        ...
Service 3 CPU      ...       ...       ...        ...
```

The goal is to make it easier to answer questions such as:

- How does CPU usage change as request rate increases?
- Does memory usage remain stable?
- Which service becomes the most resource-intensive?
- How much network traffic does a workload generate?
- At what load level does application latency start increasing?
- How does client-side performance correlate with server-side resource usage?

> **Status:** Early development / experimental

---

## Why LoadMark?

Load testing tools are very good at telling you what happened to the client:

```text
Requests
Latency
Throughput
Errors
Dropped iterations
```

Monitoring systems are very good at telling you what happened inside the infrastructure:

```text
CPU
Memory
Network
Pod health
```

The difficult part is connecting these two perspectives.

LoadMark uses the load-test stages as a common timeline:

```text
                    Benchmark
                       │
        ┌──────────────┼──────────────┐
        │              │              │
      Stage 1        Stage 2        Stage 3
      10 RPS         25 RPS         50 RPS
        │              │              │
        └───────┬──────┴──────┬───────┘
                │             │
              k6           Prometheus
                │             │
                └──────┬──────┘
                       │
                Benchmark results
```

Each stage has a start and end time. Prometheus queries use those same time windows so that infrastructure metrics can be associated with the corresponding load level.

---

## Current Features

### Load testing

- Runs k6 locally from LoadMark.
- Supports staged load profiles.
- Records benchmark start and end times.
- Records the start and end time of every configured stage.
- Streams normal k6 output to the terminal.

### Prometheus integration

- Queries Prometheus using HTTP.
- Supports authenticated Prometheus endpoints.
- Supports PromQL range queries.
- Collects metrics for specific Kubernetes namespaces and pod name patterns.
- Associates Prometheus data with individual load-test stages.
- Calculates average and peak values from collected samples.
- Supports a configurable Prometheus collection delay.

### Kubernetes resource metrics

Currently implemented:

- CPU
- Memory
- Network receive
- Network transmit

Additional resource metrics such as pod restarts, disk I/O, and database connections are part of the planned monitoring model.

---

## Architecture

LoadMark is written in Go.

The high-level flow is:

```text
                         config.yml
                             │
                             ▼
                     Configuration Loader
                             │
                 ┌───────────┴───────────┐
                 │                       │
                 ▼                       ▼
                k6                   Prometheus
                 │                       │
          ClientResult              Prometheus data
                 │                       │
                 └───────────┬───────────┘
                             │
                             ▼
                     BenchmarkResult
                             │
                             ▼
                       Terminal output
```

### Stage-based collection

A benchmark consists of one or more stages.

For example:

```yaml
stages:
  - rps: 10
    duration: 30s
  - rps: 25
    duration: 30s
  - rps: 50
    duration: 30s
  - rps: 100
    duration: 60s
```

LoadMark records the actual benchmark timeline:

```text
Stage 1
10 RPS
20:00:00 → 20:00:30

Stage 2
25 RPS
20:00:30 → 20:01:00

Stage 3
50 RPS
20:01:00 → 20:01:30

Stage 4
100 RPS
20:01:30 → 20:02:30
```

Prometheus is then queried using those stage windows.

---

# Installation

## Requirements

LoadMark currently requires:

- Go
- k6
- A Prometheus server containing the Kubernetes metrics you want to collect

### Go

Install Go from:

[https://go.dev/](https://go.dev/)

### k6

Install k6 from:

[https://grafana.com/docs/k6/latest/set-up/](https://grafana.com/docs/k6/latest/set-up/)

Verify the installation:

```bash
go version
k6 version
```

---

## Building LoadMark

Clone the repository:

```bash
git clone https://github.com/prionkor/loadmark.git
cd loadmark
```

Build:

```bash
go build -o loadmark .
```

Run:

```bash
./loadmark
```

During development, you can also run:

```bash
go run .
```

---

# Configuration

LoadMark is configured using a YAML file.

A configuration contains the following top-level sections:

```yaml
name:
description:
target:
k6:
prometheus:
monitoring:
```

A complete example:

```yaml
name: service-load-test
description: Benchmark the service under increasing load

target:
  base_url: https://app.example.com
  path: /path/to/resource
  request:
    method: POST
    headers:
      Content-Type: application/x-www-form-urlencoded
    body:
      foo: "bar"
      param2: "param 2 value"

k6:
  executor: ramping-arrival-rate
  executor_settings:
    start_rps: 10
    time_unit: 1s
    pre_allocated_vus: 10
    max_vus: 50
    stages:
      - rps: 10
        duration: 30s
      - rps: 25
        duration: 30s
      - rps: 50
        duration: 30s
      - rps: 100
        duration: 60s

prometheus:
  collection_delay: 10s
  url: https://prometheus.example.com
  auth:
    type: basic
    username:
      from_env: PROMETHEUS_USERNAME
    password:
      from_env: PROMETHEUS_PASSWORD

monitoring:
  - name: service-1
    namespace: namespace-1
    pod_name_pattern: service-1-.*
    metrics:
      cpu:
        - average
        - peak
      memory:
        - average
        - peak
      network:
        - receive
        - transmit

  - name: postgres
    namespace: namespace-2
    pod_name_pattern: pg-.*
    metrics:
      cpu:
        - average
        - peak
      memory:
        - average
        - peak

  - name: ingress
    namespace: ingress-nginx
    pod_name_pattern: ingress-nginx-controller-.*
    metrics:
      cpu:
        - average
        - peak
      memory:
        - average
        - peak
      network:
        - receive
        - transmit
```

---

# Configuration Reference

## `name`

A human-readable name for the benchmark.

```yaml
name: service-load-test
```

### Type

`string`

---

## `description`

A description of the benchmark.

```yaml
description: Benchmark the webhook service under increasing load
```

### Type

`string`

---

# Target

The `target` section describes the HTTP endpoint that k6 will test.

```yaml
target:
  base_url: https://app.example.com
  path: /path/to/resource
  request:
    method: POST
    headers:
      Content-Type: application/x-www-form-urlencoded
    body:
      foo: "bar"
      param2: "param 2 value"
```

## `target.base_url`

The base URL of the application being tested.

```yaml
base_url: https://app.example.com
```

### Type

`string`

---

## `target.path`

The request path.

```yaml
path: /path/to/resource
```

### Type

`string`

---

# Target Request

## `target.request.method`

HTTP method used for the request.

```yaml
method: POST
```

### Type

`string`

Examples:

```yaml
method: GET
```

```yaml
method: POST
```

---

## `target.request.headers`

HTTP headers sent with each request.

```yaml
headers:
  Content-Type: application/json
  Authorization: Bearer ...
```

### Type

`map[string]string`

---

## `target.request.body`

Request body parameters.

```yaml
body:
  foo: "bar"
  param2: "param 2 value"
```

### Type

`map[string]string`

---

# k6

The `k6` section controls the load test.

```yaml
k6:
  executor: ramping-arrival-rate
  executor_settings:
    start_rps: 10
    time_unit: 1s
    pre_allocated_vus: 10
    max_vus: 50
    stages:
      - rps: 10
        duration: 30s
      - rps: 25
        duration: 30s
      - rps: 50
        duration: 30s
      - rps: 100
        duration: 60s
```

## `k6.executor`

The k6 executor used by the benchmark.

```yaml
executor: ramping-arrival-rate
```

### Type

`string`

The executor must correspond to a supported k6 executor.

---

## `k6.executor_settings`

Executor-specific settings passed to the k6 test.

```yaml
executor_settings:
  start_rps: 10
  time_unit: 1s
  pre_allocated_vus: 10
  max_vus: 50
```

### Type

`map[string]any`

The exact values depend on the selected k6 executor.

---

## `k6.executor_settings.stages`

Defines the staged load profile.

```yaml
stages:
  - rps: 10
    duration: 30s

  - rps: 25
    duration: 30s

  - rps: 50
    duration: 30s

  - rps: 100
    duration: 60s
```

Each stage contains:

### `rps`

The intended request/iteration rate for the stage.

```yaml
rps: 50
```

### Type

`integer`

### `duration`

How long the stage should run.

```yaml
duration: 30s
```

### Type

`string`

The duration uses Go-style duration syntax such as:

```text
10s
30s
1m
2m30s
```

---

# Prometheus

The `prometheus` section configures the Prometheus server used for backend measurements.

```yaml
prometheus:
  collection_delay: 10s
  url: https://prometheus.example.com
  auth:
    type: basic
    username:
      from_env: PROMETHEUS_USERNAME
    password:
      from_env: PROMETHEUS_PASSWORD
```

## `prometheus.url`

Prometheus server URL.

```yaml
url: https://prometheus.example.com
```

### Type

`string`

---

## `prometheus.collection_delay`

The amount of time LoadMark waits before querying Prometheus after the benchmark has completed.

```yaml
collection_delay: 10s
```

This is necessary because Prometheus collects metrics at intervals. The most recent samples may not be available immediately when the load test finishes.

LoadMark validates this value when loading the configuration.

### Type

`string`

### Examples

```yaml
collection_delay: 5s
```

```yaml
collection_delay: 10s
```

```yaml
collection_delay: 30s
```

---

# Prometheus Authentication

LoadMark supports authentication configuration through the `auth` section.

```yaml
auth:
  type: basic
  username:
    from_env: PROMETHEUS_USERNAME
  password:
    from_env: PROMETHEUS_PASSWORD
```

## `prometheus.auth.type`

Authentication method.

```yaml
type: basic
```

### Type

`string`

---

## `prometheus.auth.username`

Prometheus username.

Values can be provided through configuration sources such as environment variables.

```yaml
username:
  from_env: PROMETHEUS_USERNAME
```

If you want to put values directly into the yaml file you can write.

```yaml
username:
  value: prom-user
```

---

## `prometheus.auth.password`

Prometheus password.

```yaml
password:
  from_env: PROMETHEUS_PASSWORD
```

### Recommended

Use environment variables rather than storing credentials directly in the YAML file.

Example:

```bash
export PROMETHEUS_USERNAME=my-user
export PROMETHEUS_PASSWORD=my-password
```

You can also add an .env file in the root. LoadMark will automatically pull those values into the environment variables.

Then:

```yaml
auth:
  type: basic
  username:
    from_env: PROMETHEUS_USERNAME
  password:
    from_env: PROMETHEUS_PASSWORD
```

---

# Monitoring

The `monitoring` section defines which Kubernetes resources LoadMark should monitor.

Example:

```yaml
monitoring:
  - name: service-1
    namespace: namespace-1
    pod_name_pattern: service-1-.*
    metrics:
      cpu:
        - average
        - peak
      memory:
        - average
        - peak
      network:
        - receive
        - transmit
```

Each monitoring resource contains:

```yaml
name:
namespace:
pod_name_pattern:
metrics:
```

---

## `monitoring[].name`

Human-readable name of the monitored resource.

```yaml
name: service-1
```

This name is used in the benchmark output.

---

## `monitoring[].namespace`

Kubernetes namespace containing the pods.

```yaml
namespace: namespace-1
```

### Type

`string`

---

## `monitoring[].pod_name_pattern`

Regular expression used to select pods.

```yaml
pod_name_pattern: service-.*
```

For example, if Kubernetes creates:

```text
service-d7c64f5df-kzpjm
```

the pattern:

```text
service-.*
```

matches the pod.

This allows LoadMark to monitor dynamically generated Kubernetes pod names without requiring Kubernetes API credentials.

### Type

`string`

### Examples

Examples below are completely arbitrary. Check your kubernetes cluster to find your pod name pattern.

```yaml
pod_name_pattern: service-.*
```

```yaml
pod_name_pattern: ingress-nginx-controller-.*
```

```yaml
pod_name_pattern: pg-.*
```

---

# Monitoring Metrics

## CPU

CPU metrics:

```yaml
cpu:
  - average
  - peak
```

Currently supported:

- `average`
- `peak`

LoadMark uses Prometheus CPU usage data and calculates these values over the stage's time window.

---

## Memory

Memory metrics:

```yaml
memory:
  - average
  - peak
```

Currently supported:

- `average`
- `peak`

Memory values are reported in human-readable units such as:

```text
8.23 MB
81.77 MB
108.34 MB
```

---

## Network

Network metrics:

```yaml
network:
  - receive
  - transmit
```

Currently supported:

- `receive`
- `transmit`

Values represent network throughput and are reported as bytes per second.

Example:

```text
Network Receive
  Average: 74.84 KB/s
  Peak:    128.15 KB/s
```

---

## Restarts

Measures how many times pods restarted during the test. The monitoring model supports:

```yaml
restarts:
  - total
```

This metric is planned for implementation.

---

## Disk

Disk monitoring is planned.

Example configuration:

```yaml
disk:
  - read
  - write
```

---

## Connections

Connection monitoring is planned for supported services.

Example:

```yaml
connections:
  - active
  - max
```

---

# Running a Benchmark

Once the configuration is ready:

```bash
./loadmark
```

or:

```bash
go run .
```

LoadMark starts the k6 benchmark and streams its output:

```text
execution: local
script: scripts/request.js

scenarios:
  * load_test: Up to 100.00 iterations/s over 4 stages
```

After the load test completes, LoadMark collects the Prometheus metrics for the configured stage windows.

---

# Example Output

The current terminal report is stage-oriented:

```text
========== BENCHMARK RESULTS ==========

Stage 1 — 10 RPS
  2026-08-31T00:42:26+06:00 → 2026-08-31T00:42:41+06:00

  webhooks
    CPU
      Average: 0.003537 cores
      Peak:    0.004391 cores
    Memory
      Average: 12.86 MB
      Peak:    12.89 MB
    Network Receive
      Average: 2269.32 B/s
      Peak:    4290.36 B/s
    Network Transmit
      Average: 1516.96 B/s
      Peak:    2852.40 B/s

  postgres
    CPU
      Average: 0.014284 cores
      Peak:    0.014322 cores
    Memory
      Average: 81.77 MB
      Peak:    81.77 MB

  ingress
    CPU
      Average: 0.009454 cores
      Peak:    0.011334 cores
    Memory
      Average: 108.13 MB
      Peak:    108.18 MB
    Network Receive
      Average: 8043.12 B/s
      Peak:    11284.37 B/s
    Network Transmit
      Average: 10173.26 B/s
      Peak:    16051.26 B/s

----------------------------------------

Stage 2 — 25 RPS
  ...
```

The stage-oriented output makes it easy to inspect what happened during a particular load level.

---

# How Prometheus Data Is Selected

LoadMark does not currently require Kubernetes API credentials to discover pods.

Instead, monitoring resources specify:

```yaml
namespace: namespace-1
pod_name_pattern: service-.*
```

LoadMark translates this into PromQL selectors.

For example:

```text
pod=~"service-.*"
```

This is combined with the configured namespace and the appropriate metric query.

This approach keeps the architecture simpler and means LoadMark only needs access to Prometheus rather than direct access to the Kubernetes API.

---

# Time Alignment

Time alignment is one of the most important parts of LoadMark.

Suppose a benchmark contains:

```text
Stage 1: 10 RPS
Stage 2: 25 RPS
Stage 3: 50 RPS
Stage 4: 100 RPS
```

LoadMark records the exact stage windows:

```text
Stage 1
Start → End

Stage 2
Start → End

Stage 3
Start → End

Stage 4
Start → End
```

Prometheus is then queried for each individual stage.

Conceptually:

```text
Client load

10 RPS   ├──────────────┤
25 RPS                  ├──────────────┤
50 RPS                                 ├──────────────┤
100 RPS                                                ├──────────────────┤
         │              │              │              │
         ▼              ▼              ▼              ▼
      Prometheus    Prometheus     Prometheus     Prometheus
      query         query          query          query
```

This allows LoadMark to compare infrastructure behavior at different load levels.

---

# Design Goals

LoadMark is intentionally designed around a few principles.

### Simple configuration

A benchmark should be describable in one YAML file.

### No Kubernetes credentials required

Prometheus already contains the Kubernetes metrics we need. LoadMark should not require direct access to the Kubernetes API just to identify pods.

### Correlated measurements

Client-side load and server-side resource usage should share the same benchmark timeline.

### Useful results

The goal is not to expose every Prometheus metric. The goal is to turn selected metrics into measurements that help answer performance questions.

### Extensible architecture

The project should be able to grow from:

```text
k6 + Prometheus
```

into a more complete benchmarking and analysis tool without coupling the entire application to either system.

---

# Current Limitations

LoadMark is still under active development.

Current limitations include:

- The report is currently terminal-based.
- Client-side k6 metrics are not yet integrated into the structured benchmark result.
- Prometheus monitoring currently focuses on CPU, memory, and network metrics.
- Restart, disk, and connection metrics are not yet fully implemented.
- The monitoring model currently relies on pod name patterns.
- The benchmark report is currently stage-oriented and does not yet provide cross-stage comparison tables or charts.
- HTML reporting is planned but not yet implemented.

---

# Roadmap

The project is evolving toward a more complete performance analysis workflow.

### Client metrics

Integrate structured k6 results into each benchmark stage, including metrics such as:

- Request count
- Error count
- Request duration
- Average latency
- P90
- P95
- P99
- Dropped iterations

### Additional Prometheus metrics

Add support for:

- Pod restarts
- Disk I/O
- Database connections
- Additional Kubernetes resource metrics

### Improved reporting

Add a self-contained HTML report containing:

- Benchmark overview
- Stage-by-stage results
- Cross-stage comparison tables
- Client-side metrics
- Server-side metrics
- Charts showing resource usage as load increases

### Analysis

Eventually, LoadMark should make it easier to identify relationships such as:

```text
Increasing RPS
      ↓
Increasing CPU
      ↓
Increasing latency
      ↓
System approaching capacity
```

---

# Project Structure

The project is written in Go and is organized around separate responsibilities.

A simplified structure:

```text
loadmark/
├── config/
├── internal/
│   ├── k6/
│   ├── metrics/
│   └── prometheus/
├── model/
├── scripts/
├── config.yml
├── go.mod
├── LICENSE
└── README.md
```

The exact project structure may evolve as the implementation grows.

---

# Prometheus Requirements

LoadMark expects Prometheus to have the Kubernetes/container metrics required by the configured monitoring resources.

For example, CPU and memory collection relies on metrics such as:

```text
container_cpu_usage_seconds_total
container_memory_working_set_bytes
```

Network collection relies on:

```text
container_network_receive_bytes_total
container_network_transmit_bytes_total
```

Your Prometheus installation must therefore be configured to scrape the appropriate Kubernetes/cAdvisor metrics.

A typical Kubernetes monitoring stack using kube-prometheus-stack can provide these metrics.

---

# Contributing

LoadMark is currently an early-stage project, and contributions, ideas, and feedback are welcome.

If you find a bug or have an idea for improving the benchmarking workflow, please open an issue.

For larger changes, opening an issue before submitting a pull request is encouraged so the proposed direction can be discussed.

---

# License

LoadMark is licensed under the **Apache License, Version 2.0**.

You may obtain a copy of the license at:

[https://www.apache.org/licenses/LICENSE-2.0](https://www.apache.org/licenses/LICENSE-2.0)

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.

See the `LICENSE` file for the complete license text.

---

# Disclaimer

LoadMark is a benchmarking and performance-analysis tool.

Benchmark results depend on the application, infrastructure, workload, Prometheus configuration, sampling interval, network conditions, and benchmark configuration.

Results should be interpreted in the context of the environment in which the benchmark was executed.
