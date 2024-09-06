---
title: Overview
expires_at: never
tags: [nfs-volume-release]
---

<!-- vim-markdown-toc GFM -->

* [Overview](#overview)
* [Running nfsv3driver in diego-cells](#running-nfsv3driver-in-diego-cells)

<!-- vim-markdown-toc -->

# Overview

Volume Services is a collection of bosh releases, service brokers, and drivers for management, installation and access to various file system types which can be mounted to a Cloud Foundry Foundation. These components exist to allow applications running on Cloud Foundry to have access to a persistent filesystem.

Volume services where implemented by extending the Open Service Broker API, enabling broker authors to create data services which have a file system based interface. Before this CF supported Shared Volumes, which is a distributed filesystems, such as NFS-based systems, which allow all instances of an application to share the same mounted volume simultaneously and access it concurrently.

Volume service added two new concepts to Cloud Foundry: Volume Mounts for Service Brokers and Volume Drivers for Diego Cells.

For more information checkout code.cloudfoundry.org/volman


# Running nfsv3driver in diego-cells

While we were discovering what all the compile-time and run-time dependencies of util-linux actually are, we discovered that we need the following two services to be running:

```
rpc.statd
rpcbind
```

When we were using debian packages for these things, we weren’t really aware that we were running these services. Now we are, we have introduced them as monit jobs. You can now customise which port statd listens on with this bosh property (which defaults to 60000, which is what the debian package was doing). You can not customise the port that rpcbind runs on, because we believe that util-linux binaries expect it to always be listening on port 111.


