# Cronjob-go

[![Keep a Changelog](https://img.shields.io/badge/changelog-Keep%20a%20Changelog-%23E05735)](CHANGELOG.md)
[![GitHub Release](https://img.shields.io/github/v/release/FMotalleb/crontab-go)](https://github.com/FMotalleb/crontab-go/releases)
[![Go Reference](https://pkg.go.dev/badge/github.com/FMotalleb/crontab-go.svg)](https://pkg.go.dev/github.com/FMotalleb/crontab-go)
[![go.mod](https://img.shields.io/github/go-mod/go-version/FMotalleb/crontab-go)](go.mod)
[![LICENSE](https://img.shields.io/github/license/FMotalleb/crontab-go)](LICENSE)
[![Build Status](https://img.shields.io/github/actions/workflow/status/FMotalleb/crontab-go/build.yml?branch=main)](https://github.com/FMotalleb/crontab-go/actions?query=workflow%3Abuild+branch%3Amain)
[![Go Report Card](https://goreportcard.com/badge/github.com/FMotalleb/crontab-go)](https://goreportcard.com/report/github.com/FMotalleb/crontab-go)
[![Codecov](https://codecov.io/gh/FMotalleb/crontab-go/branch/main/graph/badge.svg)](https://codecov.io/gh/FMotalleb/crontab-go)

⭐ `Star` this repository if you find it valuable and worth maintaining.

👁 `Watch` this repository to get notified about new releases, issues, etc.

## Description

### Cronjob-go: A Robust Cron Scheduler for Docker Environments

Cronjob-go is a powerful, lightweight, and highly configurable Golang application designed to replace the traditional crontab in Docker environments. With its seamless integration and easy-to-use YAML configuration, Cronjob-go simplifies the process of scheduling and managing recurring tasks within your containerized applications.

## Key Features

1. **YAML Configuration**: Cronjob-go leverages YAML for its configuration, making it easy to define and manage your scheduled tasks. The YAML format provides a clean and human-readable way to specify task details, schedules, and other settings.

2. **Robust Scheduling**: The application uses a reliable and flexible scheduling engine to ensure that your tasks are executed on time. It supports a wide range of scheduling patterns, including cron-style expressions, intervals, and custom scheduling logic.

3. **Container-Friendly**: Cronjob-go is designed with Containerization in mind, making it the perfect replacement for crontab in your containerized environments.

4. **Logging and Monitoring**: The application provides comprehensive logging and monitoring capabilities, allowing you to track the execution of your scheduled tasks and quickly identify and resolve any issues that may arise.

**Use Cases:**

- **Automated Backups**: Schedule regular backups of your application data or logs to ensure data integrity and disaster recovery.
- **Periodic Maintenance Tasks**: Execute maintenance tasks, such as database optimizations, cache clearance, or system updates, on a scheduled basis.
- **Data Processing and Reporting**: Automate the processing and generation of reports, analytics, or other data-driven tasks.

## Configuration

> All env variables and configuration samples can be found in `.env.example` or `config.example.yaml` files

### Logging

- **Time Format**: Uses golang's datetime format defaults to `2006-01-02T15:04:05.000Z` can be changed using `LOG_TIMESTAMP_FORMAT` env var
- **Log Format**: By default uses ansi format (colorful) can be set to (`ansi`(colorful),`plain`(no-colors) and `json`) using `LOG_FORMAT` env var
- **Log File**: Built-in support for saving logs into a file using `LOG_FILE` env var
- **StdOut**: Can be disabled using `LOG_STDOUT=false` env var
- **Log Level**: Defaults to `info` but can be set from most verbose to least being(`trace`,`debug`,`info`,`warn`,`fatal` and `panic`) using `LOG_LEVEL` env var

### Shell

In order to not making the wheel again we use your own shell to run the commands

**Shell**: You can set shell (defaults to `sh` for linux and `cmd` for windows) using `SHELL` env var (can be changed for each process explicitly)
**Shell Args**: Defaults to `-c` for `sh(linux)` and `/c` for `cmd(windows)` can be changed using `SHELL_ARGS` env var

### Config.yaml

A fully documented config file can be found in [config.example.yaml](config.example.yaml)

## Getting Started

To get started with Cronjob-go, simply download the binary for your platform and configure your scheduled tasks using the provided YAML format. The application's documentation includes detailed instructions on installation, configuration, and usage, making it easy to integrate into your existing Docker-based infrastructure.
