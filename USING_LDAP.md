# Integrating nfs-volume-release with an LDAP Server

As of version 0.1.7 it is possible to configure your deployment of `nfs-volume-release` to connect to an LDAP server.  This enables nfsv3driver to:
- Ensure that the application developer has valid credentials (according to the ldap server) to use an account.
- Translate user credentials into a valid UID and GID for that user.

The principal benefit to this feature is that it secures the nfs-volume-release so that it is no longer possible for an application develop to bind
to an NFS share using an arbitrary UID and potentially gain access to sensitive data stored by another user or application.  Once LDAP support is
enabled and regular UID and GID parameters are disabled, application developers will need to provide credentials for any user they wish to bind as.

### :bangbang: WARNING: LEAKED CREDENTIALS :bangbang:
If you are using a version of Diego before v1.12.0, the Diego Rep will leak LDAP credentials into logs at levels **info** and **debug**. We strongly recommend you install a newer version of diego, or set your rep log level to warn.

## Changes to your LDAP server
It is not generally necessary to make adjustments to your LDAP server to enable integration, but you will need the following:
- Your LDAP server must be reachable through the network from the Diego cell VMs on the port you will use to connect (normally 389 or 636)
- You should provision (or reuse) a service account on the LDAP server that has read-only access to user records.  This account will be used by 
  nfsv3driver to look up usernames and convert them to UIDs.  In Windows server 2008 or later this can be accomplished by creating a new user
  and adding it to the `Read-only Domain Controllers` group.
  
## Changes to your `nfs-volume-release` deployment.
### Broker changes
Because the actual connection to the LDAP server occurs at mount time in the volume driver, the service broker is only minimally involved in
LDAP integration.  Essentially the only change to the broker is that it should accept `username` and `password` configuration options at 
service bind time, and it should **not** accept `uid` or `gid`.  Acceptable options for the broker are specified in the following property:
- `nfsbroker.allowed_options`: Comma separated list of white-listed options that may be set during create or bind operations. 

#### `bosh deploy`ed brokers:
If you are bosh deploying your service broker, modify your manifest to set `nfsbroker.allowed_options` to `auto_cache,username,password` and redeploy.
You should see that new service bindings fail if `uid` or `gid` are specified through the `-c` option. 

#### `cf push`ed brokers:
If you are pushing your broker to cloudfoundry, the same options can be set on the command line.  Modify `Procfile` under nfsbroker to include 
`--allowedOptions="auto_cache,username,password"` in the command line arguments for nfsbroker, then `cf push` your broker app.

### Driver changes
LDAP integration in the driver is configured with the following BOSH properties:
 `nfsv3driver.allowed-in-source`: Comma separated list of white-listed options that may be configured in supported in the mount_config.source URL query params.
 `nfsv3driver.allowed-in-mount`: Comma separated list of white-listed options that may be accepted in the mount_config options. Note a specific 'sloppy_mount:true' volume option tells the driver to ignore non-white-listed options, while a 'sloppy_mount:false' tells the driver to fail fast instead when receiving a non-white-listed option."
 `nfsv3driver.ldap_svc_user`: ldap service account user name
 `nfsv3driver.ldap_svc_password`: ldap service account password
 `nfsv3driver.ldap_host`: ldap server host name or ip address
 `nfsv3driver.ldap_port`: ldap server port
 `nfsv3driver.ldap_proto`: ldap server protocol (tcp or udp)"
 `nfsv3driver.ldap_user_fqdn`: ldap fqdn for user records we will search against when looking up user uids
   
Depending on the technique you are using to deploy nfsv3driver into your release, you will need to add these properties either to your runtime-config or directly 
to your diego manifest, and then redeploy Diego.  Assuming that you are using runtime-config, your configuration should look something like this:

```yaml
-------
releases:
- name: nfs-volume
  version: 0.1.7
addons:
- name: voldrivers
  include:
    deployments: [cf-warden-diego]
    jobs: [{name: rep, release: diego}]
  jobs:
  - name: nfsv3driver
    release: nfs-volume
    properties:
      nfsv3driver:
        allowed-in-source: ""
        allowed-in-mount: auto_cache,username,password
        ldap_svc_user: readonlyuserguy
        ldap_svc_password: sup3rSecret!!!
        ldap_host: ldap.myawesomedomain.com
        ldap_port: 389
        ldap_proto: tcp
        ldap_user_fqdn: cn=Users,dc=corp,dc=myawesomedomain,dc=com
```

## Testing

Once you have redeployed your broker and driver, steps to test will be more or less the same as before, except that when you bind your 
application to your volume, you must specify `username` and `password` instead of UID and GID.  Accordingly, your bind-service command should look something like this:

```bash
$ cf bind-service pora myVolume -c '{"username":"janJansson","password":"fromW1sconson!"}'
```

Assuming that your ldap server has a user with the above credentials, and that that user has access to the share your're binding to, the app should work as before.

## Credential Rotation

Note that user credentials will be stored as part of the service binding and checked whenever an application is placed on a cell.  Hence, if the password changes, the 
application must be re-bound to the service and restaged, otherwise it will fail the next time the application is restarted or scaled.
