# NFS volume release

This is a bosh release that packages:
- an [nfsv3driver](https://github.com/cloudfoundry-incubator/nfsv3driver) 
- [nfsbroker](https://github.com/cloudfoundry-incubator/nfsbroker) 
- a test NFS server 

The broker and driver allow you to provision existing NFS volumes and bind those volumes to your applications for shared file access.

The test server provides an easy test target with which you can try out volume mounts.

# Deploying to Cloud Foundry

As of version 1.2.0 we no longer support old cf-release deployments with bosh v1 manifests.  Nfs-volume-release jobs should be added to your cf-deployment using provided ops files.

## Pre-requisites

1. Install Cloud Foundry, or start from an existing CF deployment.  If you are starting from scratch, the article [Deploying CF and Diego to AWS](https://docs.cloudfoundry.org/deploying/index.html) provides detailed instructions.

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

Your CF deployment will now have a running service broker and volume drivers, ready to mount nfs volumes.  Unless you have explicitly defined a variable for your nfsbroker password, BOSH will generate one for you.  You can find the password for use in broker registration via the `bosh interpolate` command:
    ```bash
    bosh int deployment-vars.yml --path /nfs-broker-password
    ```

If you wish to also deploy the NFS test server, you can fetch the operations file from the [persi-ci github repository](https://github.com/cloudfoundry/persi-ci/blob/master/operations/enable-nfs-test-server.yml) and include that operation with a `-o` flag also.  That will create a separate VM with nfs exports that you can use to experiment with volume mounts.
> NB: by default, the nfs test server expects that your CF deployment is deployed to a 10.x.x.x subnet.  If you are deploying to a subnet that is not 10.x.x.x (e.g. 192.168.x.x) then you will need to override the `export_cidr` property.
> Edit the generated manifest, and replace this line:
> `  nfstestserver: {}`
> with something like this:
> `  nfstestserver: {export_cidr: 192.168.0.0/16}`

# Testing or Using this Release


### Deploy the NFS Server
* Deploy the NFS server using the generated manifest:

    ```bash
    $ bosh -d nfs-test-server-aws-manifest.yml deploy
    ```

* Note the default **gid** & **uid** which are 0 and 0 respectively (root).

## Register nfs-broker
* Register the broker and grant access to its service with the following commands:

    ```bash
    $ cf create-service-broker nfsbroker <BROKER_USERNAME> <BROKER_PASSWORD> http://nfs-broker.YOUR.DOMAIN.com
    $ cf enable-service-access nfs
    ```
    Again, if you have not explicitly set a variable value for your service broker password, you can find the value bosh has assigned using the `bosh interpolate` command described above.

## Create an NFS volume service
* If you are testing against the `nfstestserver` job packaged in this release, type the following:

    ```bash
    $ cf create-service nfs Existing myVolume -c '{"share":"nfstestserver.service.cf.internal/export/vol1"}'
    $ cf services
    ```
* If you are using your own server, substitute the nfs address of your server and share, taking care to omit the `:` that ordinarily follows the server name in the address.

### NFS v4 (Experimental):

To provide our existing `nfs` service capabilities we use a libfuse implementation that only supports nfsv3 and has some performance constraints.    
  
If you require nfsv4 or better performance or both then you can try the new nfsv4 (experimental) support offered through a new nfsbroker plan called `nfs-experimental`.  

* type the following:

   ```bash
    $ cf create-service nfs-experimental Existing myVolume -c '{"share":"nfstestserver.service.cf.internal/export/vol1"}'
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
> ####Bind Parameters####
> * **uid** and **gid:** When binding the nfs service to the application, the uid and gid specified are supplied to the nfs driver.  The nfs driver tranlates the application user id and group id to the specified uid and gid when sending traffic to the nfs server, and translates this uid and gid back to the running user uid and default gid when returning attributes from the server.  This allows you to interact with your nfs server as a specific user while allowing Cloud Foundry to run your application as an arbitrary user.
> * **mount:** By default, volumes are mounted into the application container in an arbitrarily named folder under /var/vcap/data.  If you prefer to mount your directory to some specific path where your application expects it, you can control the container mount path by specifying the `mount` option.  The resulting bind command would look something like
> ``` cf bind-service pora myVolume -c '{"uid":"0","gid":"0","mount":"/var/path"}'```
> * **readonly:** Set true if you want the mounted volume to be read only. 

* Start the application
    ```bash
    $ cf start pora
    ```

## Test the app to make sure that it can access your NFS volume
* to check if the app is running, `curl http://pora.YOUR.DOMAIN.com` should return the instance index for your app
* to check if the app can access the shared volume `curl http://pora.YOUR.DOMAIN.com/write` writes a file to the share and then reads it back out again.

# Application specifics
For most buildpack applications, the workflow described above will enable NFS volume services (we have tested go, java, php and python). There are special situations to note however when using a Docker image as discussed below:

> ## Security Note
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

# Troubleshooting
If you have trouble getting this release to operate properly, try consulting the [Volume Services Troubleshooting Page](https://github.com/cloudfoundry-incubator/volman/blob/master/TROUBLESHOOTING.md)
