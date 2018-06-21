# NFS volume release

This is a bosh release that packages:
- an [nfsv3driver](https://github.com/cloudfoundry-incubator/nfsv3driver) 
- [nfsbroker](https://github.com/cloudfoundry-incubator/nfsbroker) 
- a sample NFS server with test shares
- a sample LDAP server with prepopulated accounts to match the NFS test server

The broker and driver allow you to provision existing NFS volumes and bind those volumes to your applications for shared file access.

The test NFS and LDAP servers provide easy test targets with which you can try out volume mounts.

# Deploying to Cloud Foundry

As of version 1.2.0 we no longer support old cf-release deployments with bosh v1 manifests.  Nfs-volume-release jobs should be added to your cf-deployment using provided ops files.

## Pre-requisites

1. Install Cloud Foundry, or start from an existing CF deployment.  If you are starting from scratch, the article [Overview of Deploying Cloud Foundry](https://docs.cloudfoundry.org/deploying/index.html) provides detailed instructions.

## Redeploy Cloud Foundry with nfs enabled

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
   **Note:** the above command is an example, but your deployment command should match the one you used to deploy Cloud Foundry initially, with the addition of a `-o operations/enable-nfs-volume-service.yml` option.

3. **If you are using cf-deployment version >= 2.0** then the ops file will deploy the `nfsbrokerpush` bosh errand rather than running nfsbroker as a bosh job.  You must invoke the errand to push the broker to cloud foundry where it will run as an application.
    ```bash
    $ bosh -e my-env -d cf run-errand nfs-broker-push
    ```


Your CF deployment will now have a running service broker and volume drivers, ready to mount nfs volumes.

If you wish to also deploy the NFS test server, you can include this [operations file](https://github.com/cloudfoundry/nfs-volume-release/blob/master/operations/enable-nfs-test-server.yml) with a `-o` flag also.  That will create a separate VM with nfs exports that you can use to experiment with volume mounts.
> Note: by default, the nfs test server expects that your CF deployment is deployed to a 10.x.x.x subnet.  If you are deploying to a subnet that is not 10.x.x.x (e.g. 192.168.x.x) then you will need to override the `export_cidr` property.
> Edit the generated manifest, and replace this line:
> `  nfstestserver: {}`
> with something like this:
> `  nfstestserver: {export_cidr: 192.168.0.0/16}`

# Testing or Using this Release

## Register nfs-broker
* Register the broker and grant access to its service with the following commands:

    ```bash
    $ bosh -e my-env -d cf run-errand nfs-broker-registrar
    $ cf enable-service-access nfs
    ```

## Create an NFS volume service
* If you are testing against the `nfstestserver` job packaged in this release, type the following:

    ```bash
    $ cf create-service nfs Existing myVolume -c '{"share":"nfstestserver.service.cf.internal/export/vol1"}'
    $ cf services
    ```
* If you are using your own server, substitute the nfs address of your server and share, taking care to omit the `:` that ordinarily follows the server name in the address.

### NFS v4 (Experimental):

To provide our existing `nfs` service capabilities we use a libfuse implementation that only supports nfsv3 and has some performance constraints.    

If you require nfsv4 or better performance or both then you can try the new nfsv4 (experimental) support offered through a new nfsbroker plan called `nfs-experimental`.  The `nfs-experimental` plan accepts a `version` parameter to determine which nfs protocol version to use.

* type the following:

   ```bash
    $ cf create-service nfs-experimental Existing myVolume -c '{"share":"nfstestserver.service.cf.internal/export/vol1","version":"4.1"}'
    $ cf services
    ```

## Deploy the pora test app, first by pushing the source code to CloudFoundry
* if you haven't already, clone this github repo and its submodules:

    ```bash
    $ cd ~/workspace
    $ git clone https://github.com/cloudfoundry/nfs-volume-release.git
    $ cd ~/workspace/nfs-volume-release
    $ ./scripts/update
    ```

* type the following:

    ```bash
    $ cd src/code.cloudfoundry.org/persi-acceptance-tests/assets/pora
    $ cf push pora --no-start
    ```

* Bind the service to your app supplying the correct uid and gid corresponding to what is seen on the nfs server.
    ```bash
    $ cf bind-service pora myVolume -c '{"uid":"1000","gid":"1000"}'
    ```
   > #### Bind Parameters
   > * **uid** and **gid:** When binding the nfs service to the application, the uid and gid specified are supplied to the nfs driver.  The nfs driver tranlates the application user id and group id to the specified uid and gid when sending traffic to the nfs server, and translates this uid and gid back to the running user uid and default gid when returning attributes from the server.  This allows you to interact with your nfs server as a specific user while allowing Cloud Foundry to run your application as an arbitrary user.
   > * **mount:** By default, volumes are mounted into the application container in an arbitrarily named folder under /var/vcap/data.  If you prefer to mount your directory to some specific path where your application expects it, you can control the container mount path by specifying the `mount` option.  The resulting bind command would look something like
   > ``` cf bind-service pora myVolume -c '{"uid":"0","gid":"0","mount":"/var/path"}'```
   > * **readonly:** Set true if you want the mounted volume to be read only. 
   > 
   > As of nfs-volume-release version 1.3.1, bind parameters may also be specified in configuration during service instance creation.  Specifying bind parameters in advance when creating the service instance is particularly helpful when binding services to an application in the application manifest, where bind configuration is not supported.

* Start the application
    ```bash
    $ cf start pora
    ```

## Test the app to make sure that it can access your NFS volume
* to check if the app is running, `curl http://pora.YOUR.DOMAIN.com` should return the instance index for your app
* to check if the app can access the shared volume `curl http://pora.YOUR.DOMAIN.com/write` writes a file to the share and then reads it back out again.

> # Security Note
> Because connecting to NFS shares will require you to open your NFS mountpoint to all Diego cells, and outbound traffic from application containers is NATed to the Diego cell IP address, there is a risk that an application could initiate an NFS IP connection to your share and gain unauthorized access to data.
> 
> To mitigate this risk, consider one or more of the following steps:
> * Avoid using `insecure` NFS exports, as that will allow non-root users to connect on port 2049 to your share.
> * Avoid enabling Docker application support as that will allow root users to connect on port 111 even when your share is not `insecure`.
> * Use [CF Security groups](https://docs.cloudfoundry.org/adminguide/app-sec-groups.html) to block direct application access to your NFS server IP, especially on ports 111 and 2049.

## File Locking via flock() and lockf()/fcntl()
If your application relies on file locking either through unix system calls such as flock() and fcntl() or through script commands such as `flock` **please be aware that the lock will not be enforced across diego cells**.  This is because the file locking implementations in the underlying fuse-nfs executable are not implemented, so locking is limited to local locks between precesses on the same VM.  If you have a legitimate requirement for file locking, please document your use case in a comment on [this github issue](https://github.com/cloudfoundry-incubator/nfs-volume-release/issues/13) and we'll see what we can do.

# LDAP Support
For better security, it is recommended to configure your deployment of nfs-volume-release to connect to an external LDAP server to resolve user credentials into uids.  See [this note](USING_LDAP.md) for more details.

# BBR Support
If you are using [Bosh Backup and Restore](https://docs.cloudfoundry.org/bbr/) (BBR) to keep backups of your Cloud Foundry deployment, consider including the [enable-nfs-broker-backup.yml](https://github.com/cloudfoundry/cf-deployment/blob/master/operations/experimental/enable-nfs-broker-backup.yml) operations file from cf-deployment when you redeploy Cloud Foundry.  This file will install the requiste backup and restore scripts for nfs service broker metadata on the backup/restore VM.

# (Experimental) Support for PXC databases
If you plan to enable the [PXC database](https://github.com/cloudfoundry/cf-deployment/blob/master/operations/experimental/use-pxc.yml) in your Cloud Foundry deployment, then you will need to apply the following ops file to allow the nfs broker to connect to PXC instead of MySql:
- [use-pxc-for-nfs-broker.yml](https://github.com/cloudfoundry/nfs-volume-release/blob/master/operations/use-pxc-for-nfs-broker.yml)

Note that because PXC enables TLS using a server certification, nfs broker will no longer be able to connect to it using an IP address.  As a result, you must also apply ops files to enable BOSH DNS, and to apply BOSH DNS to application containers in order to allow the nfs broker to connect to PXC using a host name:
- [use-bosh-dns.yml](https://github.com/cloudfoundry/cf-deployment/blob/master/operations/experimental/use-bosh-dns.yml)
- [use-bosh-dns-for-containers.yml](https://github.com/cloudfoundry/cf-deployment/blob/master/operations/experimental/use-bosh-dns-for-containers.yml)

# (Experimental) Support for CredHub as a backing store for nfs broker

Version 1.4.0 introduces support for using CredHub instead of a SQL database to store state for nfs broker.  CredHub has the advantage that it encrypts data at rest and is therefore a more secure store for service instance and service binding metadata.  CredHub is required if you are using the LDAP integration, and you wish to specify user credentials at service instance creation time, rather than at service binding time.  To use CredHub as the backing store for nfs broker, apply this ops file:
- [enable-nfs-volume-service-credhub.yml](https://github.com/cloudfoundry/nfs-volume-release/blob/master/operations/enable-nfs-volume-service-credhub.yml)

Note that this ops file will install a separate errand for the credhub enabled broker.  To push that broker and register it you should type the following:

```bash
$ bosh -e my-env -d cf run-errand nfs-broker-credhub-push
$ bosh -e my-env -d cf run-errand nfs-broker-credhub-registrar
```

# Troubleshooting
If you have trouble getting this release to operate properly, try consulting the [Volume Services Troubleshooting Page](https://github.com/cloudfoundry-incubator/volman/blob/master/TROUBLESHOOTING.md)
