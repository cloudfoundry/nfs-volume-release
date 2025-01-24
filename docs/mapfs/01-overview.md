---
title: Overview
expires_at: never
tags: [nfs-volume-release]
---

<!-- vim-markdown-toc GFM -->

* [Overview](#overview)
    * [MapFS](#mapfs)
    * [Performance](#performance)
    * [Configuration](#configuration)
* [Usage](#usage)

<!-- vim-markdown-toc -->
# Overview

## MapFS
Go-based [FUSE](https://en.wikipedia.org/wiki/Filesystem_in_Userspace) file system for uid mapping local file access


In CloudFoundry [NFS volume services](https://github.com/cloudfoundry/nfs-volume-release),
the [NFS driver](https://github.com/cloudfoundry/nfsv3driver) will start `mapfs` when the `uid` or `gid` parameters
are specified when binding an NFS mount. MapFS will expose a FUSE filesystem that is mounted into the app container,
and the app can read and write from. This is backed by an NFS mount.
MapFS replicates file access so that as far as NFS is concerned, the file access is done
with the `uid` and `gid` specified.

## Performance
Because an app can potentially go faster than NFS (which is limited by network bandwidth), it's possible for requests
to stack up in MapFS while it waits for NFS to respond. Currently, [go-fuse](https://github.com/hanwen/go-fuse) does
not implement backpressure or load shedding. For every request in-flight, a buffer and a goroutine will be created,
so MapFS can consume high memory and processor when the app accesses the filesystem faster than NFS responds.

## Configuration
There is an optional configuration file that can be placed at `/var/vcap/jobs/map-fs/config/mapfs.yml` that will allow
MapFS to be configured. Configuration keys are:
- `debug` - enables go-fuse debug mode. Logs are written to `/var/vcap/sys/log/map-fs/mapfs.$$.log` where `$$` is the process ID.
- `single_threaded` - enables go-fuse single threaded mode. Can reduce memory pressure, and may slow down access.
- `cpu_profile` - writes a Go pprof CPU profile to the specified path.
- `mem_profile` - when the process receives a `SIGUSR1` a Go pprof memory profile will be written to the path with a timestamp as a file suffix.
- `soft_mem_limit` - soft memory limit in bytes. Equivalent to setting `GOMEMLIMIT` environment variable.

The configuration file is read on startup. To restart MapFS just `kill -9` the existing
process and the NFS Driver will start a new one. This may break running apps.


# Usage

map-fs job is added to cf-deployment when using [enable-nfs-volume-service.yml](https://github.com/cloudfoundry/cf-deployment/blob/main/operations/enable-nfs-volume-service.yml)


