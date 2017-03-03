# NFS volume release
This is a bosh release that packages an [nfsv3driver](https://github.com/cloudfoundry-incubator/nfsv3driver), [nfsbroker](https://github.com/cloudfoundry-incubator/nfsbroker) and a test NFS server for consumption by a volume_services_enabled Cloud Foundry deployment.

This broker/driver pair allows you to provision existing NFS volumes and bind those volumes to your applications for shared file access.

# Deploying to AWS EC2

## Pre-requisites

1. Install Cloud Foundry with Diego, or start from an existing CF+Diego deployment on AWS.  If you are starting from scratch, the article [Deploying CF and Diego to AWS](https://github.com/cloudfoundry/diego-release/tree/develop/examples/aws) provides detailed instructions.

2. If you don't already have it, install spiff according to its [README](https://github.com/cloudfoundry-incubator/spiff). spiff is a tool for generating BOSH manifests that is required in some of the scripts used below.

## Create and Upload this Release

1. Check out nfs-volume-release (master branch) from git:

    ```bash
    $ cd ~/workspace
    $ git clone https://github.com/cloudfoundry-incubator/nfs-volume-release.git
    $ cd ~/workspace/nfs-volume-release
    $ git checkout master
    $ ./scripts/update
    ```

2. Bosh Create and Upload the release
    ```bash
    $ bosh -n create release --force && bosh -n upload release
    ```

## Enable Volume Services in CF and Redeploy

In your CF manifest, check the setting for `properties: cc: volume_services_enabled`.  If it is not already `true`, set it to `true` and redeploy CF.  (This will be quick, as it only requires BOSH to restart the cloud controller job with the new property.)

## Colocate the nfsv3driver job on the Diego Cell
If you have a bosh director version < `259` you will need to use one of the OLD WAYS below. (check `bosh status` to determine your version).  Otherwise we recommend the NEW WAY :thumbsup::thumbsup::thumbsup:

### OLD WAY #1 Using Scripts to generate the Diego Manifest
If you originally created your Diego manifest from the scripts in diego-release, then you can use the same scripts to recreate the manifest with `nfsv3driver` included.

1. In your diego-release folder, locate the file `manifest-generation/bosh-lite-stubs/experimental/voldriver/drivers.yml` and copy it into your local directory.  Edit it to look like this:

    ```yaml
    volman_overrides:
      releases:
      - name: nfs-volume
        version: "latest"
      driver_templates:
      - name: nfsv3driver
        release: nfs-volume
    ```

2. Now regenerate your diego manifest using the `-d` option, as detailed in [Setup Volume Drivers for Diego](https://github.com/cloudfoundry/diego-release/blob/develop/examples/aws/OPTIONAL.md#setup-volume-drivers-for-diego)

3. Redeploy Diego.  Again, this will be a fast operation as it only needs to start the new `nfsv3driver` job on each Diego cell.

### OLD WAY #2 Manual Editing
If you did not use diego scripts to generate your manifest, you can manually edit your diego manifest to include the driver.

1. Add `nfs-volume` to the `releases:` key

    ```yaml
    releases:
    - name: diego
      version: latest
      ...
    - name: nfs-volume
      version: latest
    ```
2. Add `nfsv3driver` to the `jobs: name: cell_z1 templates:` key

    ```yaml
    jobs:
      ...
      - name: cell_z1
        ...
        templates:
        - name: consul_agent
          release: cf
          ...
        - name: nfsv3driver
          release: nfs-volume
    ```

3. If you are using multiple AZz, repeat step 2 for `cell_z2`, `cell_z3`, etc.

4. Redeploy Diego using your new manifest.

### NEW WAY Use bosh add-ons with filtering
This technique allows you to co-locate bosh jobs on cells without editing the Diego bosh manifest.

1. Create a new `runtime-config.yml` with the following content:

    ```yaml
    ---
    releases:
    - name: nfs-volume
      version: <YOUR VERSION HERE>
    addons:
    - name: voldrivers
      include:
        deployments:
        - <YOUR DIEGO DEPLOYMENT NAME>
        jobs:
        - name: rep
          release: diego
      jobs:
      - name: nfsv3driver
        release: nfs-volume
        properties: {}
    ```

2. Set the runtime config, and redeploy diego

    ```bash
    $ bosh update runtime-config runtime-config.yml
    $ bosh download manifest <YOUR DIEGO DEPLOYMENT NAME> diego.yml
    $ bosh -d diego.yml deploy
    ```

## Deploying nfsbroker

The nfsbroker can be deployed in two ways; as a cf app or as a BOSH deployment.  The choice is yours!

### Way #1 `cf push` the broker

When the service broker is `cf push`ed, it must be bound to a MySql or Postgres database service instance.  (Since Cloud Foundry applications are stateless, it is not safe to store state on the local filesystem, so we require a database to do simple bookkeeping.)

Once you have a database service instance available in the space where you will push your service broker application, follow the following steps:
- `cd src/code.cloudfoundry.org/nfsbroker`
- `GOOS=linux GOARCH=amd64 go build -o bin/nfsbroker`
- edit `manifest.yml` to set up broker username/password and sql db driver name and cf service name.  If you are using the [cf-mysql-release](http://bosh.io/releases/github.com/cloudfoundry/cf-mysql-release) from bosh.io, then the database parameters in manifest.yml will already be correct.
- `cf push <broker app name> --no-start`
- `cf bind-service <broker app name> <sql service instance name>`
- `cf start <broker app name>`

### Way #2 - `bosh deploy` the broker

#### Create Stub Files

##### cf.yml

* copy your cf.yml that you used during cf deployment, or download it from bosh: `bosh download manifest [your cf deployment name] > cf.yml`

##### director.yml
* determine your bosh director uuid by invoking bosh status --uuid
* create a new director.yml file and place the following contents into it:

    ```yaml
    ---
    director_uuid: <your uuid>
    ```

##### iaas.yml

* Create a stub for your iaas settings from the following template:

    ```yaml
        ---
        jobs:
        - name: nfsbroker
          networks:
          - name: public
            static_ips: [<--- STATIC IP WANT YOUR NFSBROKER TO BE IN --->]

        networks:
        - name: nfsvolume-subnet
          subnets:
          - cloud_properties:
              security_groups:
              - <--- SECURITY GROUP YOU WANT YOUR NFSBROKER TO BE IN --->
              subnet: <--- SUBNET YOU WANT YOUR NFSBROKER TO BE IN --->
            dns:
            - 10.10.0.2
            gateway: 10.10.200.1
            range: 10.10.200.0/24
            reserved:
            - 10.10.200.2 - 10.10.200.9
            # ceph range 10.10.200.106-110
            # local range 10.10.200.111-115
            # efs range 10.10.200.116-120
            - 10.10.200.106 - 10.10.200.120
            # -> nfs range 10.10.200.121-125 <-
            static:
            - 10.10.200.10 - 10.10.200.105

        resource_pools:
          - name: medium
            stemcell:
              name: bosh-aws-xen-hvm-ubuntu-trusty-go_agent
              version: latest
            cloud_properties:
              instance_type: m3.medium
              availability_zone: us-east-1c
          - name: large
            stemcell:
              name: bosh-aws-xen-hvm-ubuntu-trusty-go_agent
              version: latest
            cloud_properties:
              instance_type: m3.large
              availability_zone: us-east-1c
    ```

NB: manually edit to fix hard-coded ip ranges, security groups and subnets to match your deployment.

##### properties.yml
* Minimally determine the following information:

    - BROKER_USERNAME: some invented username
    - BROKER_PASSWORD: some invented password

* create a new properties.yml file and place the following contents into it:
    ```yaml
        ---
        properties:
          nfsbroker:
            username: <BROKER_USERNAME>
            password: <BROKER_PASSWORD>
    ```

* optionally you can add other properties here:
    ```config
     nfsbroker.listen_addr:
       description: "(optional) address nfsbroker listens on"
       default: "0.0.0.0:8999"
     nfsbroker.service_name:
       description: "(optional) name of the service to be registered with cf"
       default: "nfs"
     nfsbroker.service_id:
       description: "(optional) Id of the service to be registered with cf"
       default: "nfs-service-guid"
     nfsbroker.data_dir:
       description: "(optional) Directory on broker VM to persist instance and binding state"
       default: "/var/vcap/store/nfsbroker"
     nfsbroker.db_driver:
       default: ""
       description: "(optional) database driver name when using SQL to store broker state"
     nfsbroker.db_username:
       default: ""
       description: "(optional) database username when using SQL to store broker state"
     nfsbroker.db_password:
       default: ""
       description: "(optional) database password when using SQL to store broker state"
     nfsbroker.db_hostname:
       default: ""
       description: "(optional) database hostname when using SQL to store broker state"
     nfsbroker.db_port:
       default: ""
       description: "(optional) database port when using SQL to store broker state"
     nfsbroker.db_name:
       default: ""
       description: "(optional) database name when using SQL to store broker state"
     nfsbroker.db_ca_cert:
       default: ""
       description: "(optional) CA Cert to verify SSL connection, if not include, connection will be plain"
    ```

    * For example: for a secure mysql database, properties.yml could look like:
    ```yaml
        ---
        properties:
          nfsbroker:
            username: <BROKER_USERNAME>
            password: <BROKER_PASSWORD>
            db_driver: mysql
            db_username: <DATABASE_USERNAME>
            db_password: <DATABASE_PASSWORD>
            db_hostname: mysql.example.com
            db_port: 3306
            db_name: mysql-example
            db_ca_cert: |
                -----BEGIN CERTIFICATE-----
                MIID9DCCAtygAwIBAgIBQjANBgkqhkiG9w0BAQ<...>VMx
                EzARBgNVBAgMCldhc2hpbmd0b24xEDAOBgNVBA<...>AoM
                GUFtYXpvbiBXZWIgU2VydmljZXMsIEluYy4xEz<...>FMx
                GzAZBgNVBAMMEkFtYXpvbiBSRFMgUm9vdCBDQT<...>w0y
                MDAzMDUwOTExMzFaMIGKMQswCQYDVQQGEwJVUz<...>3Rv
                bjEQMA4GA1UEBwwHU2VhdHRsZTEiMCAGA1UECg<...>WNl
                cywgSW5jLjETMBEGA1UECwwKQW1hem9uIFJEUz<...>FJE
                UyBSb290IENBMIIBIjANBgkqhkiG9w0BAQEFAA<...>Z8V
                u+VA8yVlUipCZIKPTDcOILYpUe8Tct0YeQQr0u<...>HgF
                Ji2N3+39+shCNspQeE6aYU+BHXhKhIIStt3r7g<...>Arf
                AOcjZdJagOMqb3fF46flc8k2E7THTm9Sz4L7RY<...>Ob9
                T53pQR+xpHW9atkcf3pf7gbO0rlKVSIoUenBlZ<...>I2J
                P/DSMM3aEsq6ZQkfbz/Ilml+Lx3tJYXUDmp+Zj<...>vwp
                BIOqsqVVTvw/CwIDAQABo2MwYTAOBgNVHQ8BAf<...>AUw
                AwEB/zAdBgNVHQ4EFgQUTgLurD72FchM7Sz1Bc<...>oAU
                TgLurD72FchM7Sz1BcGPnIQISYMwDQYJKoZIhv<...>pAm
                MjHD5cl6wKjXxScXKtXygWH2BoDMYBJF9yfyKO<...>Aw5
                2EUuDI1pSBh9BA82/5PkuNlNeSTB3dXDD2PEPd<...>m4r
                47QPyd18yPHrRIbtBtHR/6CwKevLZ394zgExqh<...>pjf
                2u6O/+YE2U+qyyxHE5Wd5oqde0oo9UUpFETJPV<...>kIV
                A9dY7IHSubtCK/i8wxMVqfd5GtbA8mmpeJFwnD<...>UYr
                /40NawZfTUU=
                -----END CERTIFICATE-----
    ```
* Other notes:
    > For previously deployed nfs brokers without databases:
        When you deploy with a database the current state will be lost.
        This will require a manual cleanup of any existing broker/service
        instances in your CF environment. You may need to force things:
    ```bash
    cf purge-service-instance -f $(cf services | grep nfs* | awk '{print $1}')
    cf purge-service-offering nfs -f
    cf delete-service-broker nfsbroker -f
    ```
    > Stay tuned for the CF pushable version of this broker.
#### Generate the Deployment Manifest
* run the following script:

    ```bash
    $ ./scripts/generate_manifest.sh cf.yml director-uuid.yml iaas.yml properties.yml
    ```

to generate `nfsvolume-aws-manifest.yml` into the current directory.

#### Deploy NFS Broker
* Deploy the broker using the generated manifest:

    ```bash
    $ bosh -d nfsvolume-aws-manifest.yml deploy
    ```

# Testing or Using this Release

## Deploying the Test NFS Server (Optional)

If you do not have an existing NFS Server then you can optionally deploy the test nfs server bundled in this release.

### Generate the Deployment Manifest

#### Create Stub Files

##### director.yml
* determine your bosh director uuid by invoking bosh status --uuid
* create a new director.yml file and place the following contents into it:

    ```yaml
    ---
    director_uuid: <your uuid>
    ```

#### iaas.yml

* Create a stub for your iaas settings from the following template:

    ```yaml
        ---
        networks:
        - name: nfsvolume-subnet
          subnets:
          - cloud_properties:
              security_groups:
              - <--- SECURITY GROUP YOU WANT YOUR NFSBROKER TO BE IN --->
              subnet: <--- SUBNET YOU WANT YOUR NFSBROKER TO BE IN --->
            dns:
            - 10.10.0.2
            gateway: 10.10.200.1
            range: 10.10.200.0/24
            reserved:
            - 10.10.200.2 - 10.10.200.9
            # ceph range 10.10.200.106-110
            # local range 10.10.200.111-115
            # efs range 10.10.200.116-120
            # nfs range 10.10.200.121-125
            - 10.10.200.106 - 10.10.200.125
            static:
            - 10.10.200.10 - 10.10.200.105

        resource_pools:
          - name: medium
            stemcell:
              name: bosh-aws-xen-hvm-ubuntu-trusty-go_agent
              version: latest
            cloud_properties:
              instance_type: m3.medium
              availability_zone: us-east-1c
          - name: large
            stemcell:
              name: bosh-aws-xen-hvm-ubuntu-trusty-go_agent
              version: latest
            cloud_properties:
              instance_type: m3.large
              availability_zone: us-east-1c

        nfs-test-server:
          ips: [<--- PRIVATE IP ADDRESS --->]
          public_ips: [<--- PUBLIC IP ADDRESS --->]
    ```

NB: manually edit to fix hard-coded ip ranges, security groups and subnets to match your deployment.

* run the following script:

    ```bash
    $ ./scripts/generate_server_manifest.sh director-uuid.yml iaas.yml
    ```

to generate `nfs-test-server-aws-manifest.yml` into the current directory.

### Deploy the NFS Server
* Deploy the NFS server using the generated manifest:

    ```bash
    $ bosh -d nfs-test-server-aws-manifest.yml deploy
    ```

* Note the default **gid** & **uid** which are 0 and 0 respectively (root).

## Register nfs-broker
* Register the broker and grant access to it's service with the following command:

    ```bash
    $ cf create-service-broker nfsbroker <BROKER_USERNAME> <BROKER_PASSWORD> http://nfs-broker.YOUR.DOMAIN.com
    $ cf enable-service-access nfs
    ```

## Create an NFS volume service
* type the following:

    ```bash
    $ cf create-service nfs Existing myVolume -c '{"share":"<PRIVATE_IP>/export/vol1"}'
    $ cf services
    ```

## Deploy the pora test app, first by pushing the source code to CloudFoundry
* type the following:

    ```bash
    $ cd src/code.cloudfoundry.org/persi-acceptance-tests/assets/pora

    $ cf push pora --no-start
    ```

* Bind the service to your app supplying the correct uid and gid corresponding to what is seen on the nfs server.
    ```bash
    $ cf bind-service pora myVolume -c '{"uid":"0","gid":"0"}'
    ```
> ####Bind Parameters####
> * **uid & gid:** When binding the nfs service to the application, the uid and gid specified are supplied to the fuse-nfs driver.  The fuse-nfs driver acts as a middle layer (translation table) to mask the running user id and group id as the true owner shown on the nfs server.  Any operation on the mount will be executed as the owner, but locally the mount will be seen as being owned by the running user.
> * **mount:** By default, volumes are mounted into the application container in an arbitrarily named folder under /var/vcap/data.  If you prefer to mount your directory to some specific path where your application expects it, you can control the container mount path by specifying the `mount` option.  The resulting bind command would look something like
> ``` cf bind-service pora myVolume -c '{"uid":"0","gid":"0","mount":"/my/path"}'```

* Start the application
    ```bash
    $ cf start pora
    ```

## Test the app to make sure that it can access your NFS volume
* to check if the app is running, `curl http://pora.YOUR.DOMAIN.com` should return the instance index for your app
* to check if the app can access the shared volume `curl http://pora.YOUR.DOMAIN.com/write` writes a file to the share and then reads it back out again.

# Application specifics
For most buildpack applications, the workflow described above will enable NFS volume services (we have tested go, java, php and python). There are special situations to note however when using a Docker image as discussed below:

## Special notes for Docker Image based apps
The user running the application inside the docker image must either have uid 0 and gid 0 (This is the root user and default docker user), or have uid 2000 and gid 2000. Below is a table showcasing what we have tested with success and failures.

| uid:gid | Description | Result |
|:----------|:-------------|:-----|
| 2000:2000 | Any CF Buildpack Default User -- CVCAP User | Success |
| 0:0 | Docker Default User -- Root User | Success |
| 20:20 | Custom User Created | Failure |

> ## Security Note
> Because connecting to NFS shares will require you to open your NFS mountpoint to all Diego cells, and outbound traffic from application containers is NATed to the Diego cell IP address, there is a risk that an application could initiate an NFS IP connection to your share and gain unauthorized access to data.
> 
> To mitigate this risk, consider one or more of the following steps:
> * Avoid using `insecure` NFS exports, as that will allow non-root users to connect on port 2049 to your share.
> * Avoid enabling Docker application support as that will allow root users to connect on port 111 even when your share is not `insecure`.
> * Use [CF Security groups](https://docs.cloudfoundry.org/adminguide/app-sec-groups.html) to block direct application access to your NFS server IP, especially on ports 111 and 2049.

# Troubleshooting
If you have trouble getting this release to operate properly, try consulting the [Volume Services Troubleshooting Page](https://github.com/cloudfoundry-incubator/volman/blob/master/TROUBLESHOOTING.md)
