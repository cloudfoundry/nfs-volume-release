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

Please note that the following steps will not work with bosh-lite.

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
> Security note: because connecting to NFS shares will require you to open your NFS mountpoint to all Diego cells, and outbound traffic from application containers is NATed to the Diego cell IP address, there is a risk that an application could initiate an NFS IP connection to your share and gain unauthorized access to data.
> 
> To mitigate this risk, consider one or more of the following steps:
> * Avoid using `insecure` NFS exports, as that will allow non-root users to connect on port 2049 to your share.
> * Avoid enabling Docker application support as that will allow root users to connect on port 111 even when your share is not `insecure`.
> * Use [CF Security groups](https://docs.cloudfoundry.org/adminguide/app-sec-groups.html) to block direct application access to your NFS server IP, especially on ports 111 and 2049.

### Test NFS Server
If you wish to also deploy the NFS test server, you can include this [operations file](https://github.com/cloudfoundry/cf-deployment/blob/master/operations/test/enable-nfs-test-server.yml) with a `-o` flag also.  That will create a separate VM with nfs exports that you can use to experiment with volume mounts.
> Note: by default, the nfs test server expects that your CF deployment is deployed to a 10.x.x.x subnet.  If you are deploying to a subnet that is not 10.x.x.x (e.g. 192.168.x.x) then you will need to override the `export_cidr` property.
> Edit the generated manifest, and replace this line:
> `  nfstestserver: {}`
> with something like this:
> `  nfstestserver: {export_cidr: 192.168.0.0/16}`

# Registering nfs-broker
* Register the broker and grant access to its service with the following commands:

    ```bash
    $ bosh -e my-env -d cf run-errand nfs-broker-registrar
    $ cf enable-service-access nfs
    ```

# Testing and General Usage

You can refer to the [Cloud Foundry docs](https://docs.cloudfoundry.org/devguide/services/using-vol-services.html#-nfs-volume-service) for testing and general usage information.

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
