---
title: Deploying to CloudFoundry
expires_at: never
tags: [nfs-volume-release]
---

<!-- vim-markdown-toc GFM -->

* [Deploying to CloudFoundry](#deploying-to-cloudfoundry)
    * [Pre-requisites](#pre-requisites)
    * [Redeploy Cloud Foundry with NFS enabled](#redeploy-cloud-foundry-with-nfs-enabled)
* [Testing or Using this Release](#testing-or-using-this-release)
    * [Deploying the Test NFS Server (Optional)](#deploying-the-test-nfs-server-optional)
    * [Register nfsbroker](#register-nfsbroker)
    * [Testing and General Usage with nfsbroker](#testing-and-general-usage-with-nfsbroker)
    * [Follow the cf docs to deploy and test a sample app](#follow-the-cf-docs-to-deploy-and-test-a-sample-app)

<!-- vim-markdown-toc -->

# Deploying to CloudFoundry

## Pre-requisites

1. Install Cloud Foundry, or start from an existing CF deployment.  If you are starting from scratch, the article 
    [Overview of Deploying Cloud Foundry](https://docs.cloudfoundry.org/deploying/index.html) provides detailed
    instructions.

## Redeploy Cloud Foundry with NFS enabled

1. You should have it already after deploying Cloud Foundry, but if not clone the cf-deployment repository from git:

    ```bash
    $ cd ~/workspace
    $ git clone https://github.com/cloudfoundry/cf-deployment.git
    $ cd ~/workspace/cf-deployment
    ```

2. Now redeploy your cf-deployment while including the nfs ops file:
    
   ```bash
    $ bosh -e my-env -d cf deploy cf.yml -v deployment-vars.yml -o operations/enable-nfs-volume-service.yml
    ```
    
> [!NOTE]
> The above command is an example, but your deployment command should match the one you used to deploy Cloud 
Foundry initially, with the addition of a `-o operations/enable-nfs-volume-service.yml` option.

> [!NOTE]
> If you'd like to run with ldap server, also include `-o operations/enable-nfs-ldap.yml` opsfile.

Your CF deployment will now have a running service broker and volume drivers, ready to mount or create NFS volumes.  
Unless you have explicitly defined a variable for your broker password, BOSH will generate one for you.

# Testing or Using this Release

## Deploying the Test NFS Server (Optional)

If you do not have an existing NFS Server then you can optionally deploy the test NFS server bundled in this release.

The easiest way to deploy the test server is to include the `enable-nfs-test-server.yml` operations file when you deploy
Cloud Foundry:

   ```bash
   $ bosh -e my-env -d cf deploy cf.yml -v deployment-vars.yml \
     -o operations/enable-nfs-volume-service.yml \
     -o operations/test/enable-nfs-test-server.yml
   ```

After deploying, test server can be reached at `nfstestserver.service.cf.internal`. 

> [!NOTE]
> If you'd like to add ldap test server, also include `-o operations/test/enable-nfs-test-ldapserver.yml`

> [!IMPORTANT]
> In order for containers to reach nfstestserver, you will also need a CF security-group that will allow containers to reach the server.

## Register nfsbroker

* Deploy and register the broker and grant access to its service with the following command:

    ```bash
    $ bosh -e my-env -d cf run-errand nfsbrokerpush
    $ cf enable-service-access nfs
    ```

## Testing and General Usage with nfsbroker

You can refer to the [Cloud Foundry docs](https://docs.cloudfoundry.org/devguide/services/using-vol-services.html#nfs) 
for testing and general usage information.

## Follow the cf docs to deploy and test a sample app

Test instructions are [here](https://docs.cloudfoundry.org/devguide/services/using-vol-services.html#nfs-sample)
The nfsbroker uses credhub as a backing store, and as a result, does not require separate scripts for backup and restore,
since credhub itself will get backed up by BBR.

