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

### Create Stub Files

#### cf.yml

* copy your cf.yml that you used during cf deployment, or download it from bosh: `bosh download manifest [your cf deployment name] > cf.yml`

#### director.yml 
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

#### creds.yml
* Determine the following information
    - BROKER_USERNAME: some invented username 
    - BROKER_PASSWORD: some invented password
    
* create a new creds.yml file and place the following contents into it:
    ```yaml
        ---
        properties:
          nfsbroker:
            username: <BROKER_USERNAME>
            password: <BROKER_PASSWORD>
    ```

### Generate the Deployment Manifest
* run the following script:

    ```bash
    $ ./scripts/generate_manifest.sh cf.yml director-uuid.yml iaas.yml creds.yml  
    ```

to generate `nfsvolume-aws-manifest.yml` into the current directory.

### Deploy NFS Broker
* Deploy the broker using the generated manifest: 

    ```bash
    $ bosh -d nfsvolume-aws-manifest.yml deploy
    ```
   
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

that will generate `nfs-test-server-aws-manifest.yml` in the current directory.

### Deploy the NFS Server
* type the following: 
    ```bash
    $ bosh -d nfs-test-server-aws-manifest.yml deploy
    ```
    
## Register nfs-broker
* type the following: 
    ```bash
    $ cf create-service-broker nfsbroker <BROKER_USERNAME> <BROKER_PASSWORD> http://nfs-broker.YOUR.DOMAIN.com
    $ cf enable-service-access nfs
    ```

## Create an NFS volume service
* type the following: 
    ```bash
    $ cf create-service nfs Existing myVolume -c '{"share":"<PRIVATE_IP>/export/users"}'
    $ cf services
    ```

## Deploy the pora test app, bind it to your service and start the app
* type the following:
 
    ```bash
    $ cd src/code.cloudfoundry.org/persi-acceptance-tests/assets/pora
    
    $ cf push pora --no-start
    
    $ cf bind-service pora myVolume
    
    $ cf start pora
    ```

## Test the app to make sure that it can access your NFS volume
* to check if the app is running, `curl http://pora.YOUR.DOMAIN.com` should return the instance index for your app
* to check if the app can access the shared volume `curl http://pora.YOUR.DOMAIN.com/write` writes a file to the share and then reads it back out again.
* test files will be written as the application user 1000:1000

