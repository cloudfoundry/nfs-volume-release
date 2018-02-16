# Integrating nfs-volume-release with an LDAP Server

For better data security, it is recommended to configure your deployment of `nfs-volume-release` to connect to an LDAP server.  This enables nfsv3driver to:
- Ensure that the application developer has valid credentials (according to the ldap server) to use an account.
- Translate user credentials into a valid UID and GID for that user.

The principal benefit to this feature is that it secures the nfs-volume-release so that it is no longer possible for an application developer to bind
to an NFS share using an arbitrary UID and potentially gain access to sensitive data stored by another user or application.  Once LDAP support is
enabled and regular UID and GID parameters are disabled, application developers will need to provide credentials for any user they wish to use on the nfs server.

## Changes to your LDAP server
It is not generally necessary to make adjustments to your LDAP server to enable integration, but you will need the following:
- Your LDAP server must be reachable through the network from the Diego cell VMs on the port you will use to connect (normally 389 or 636)
- You should provision (or reuse) a service account on the LDAP server that has read-only access to user records.  This account will be used by 
  nfsv3driver to look up usernames and convert them to UIDs.  In Windows server 2008 or later this can be accomplished by creating a new user
  and adding it to the `Read-only Domain Controllers` group.
  
## Changes to your `nfs-volume-release` deployment.
Assuming that you have used the `enable-nfs-volume-service.yml` operations file to include `nfs-volume-release` in your deployment, you can use the
[`enable-nfs-ldap`](https://github.com/cloudfoundry/cf-deployment/blob/master/operations/enable-nfs-ldap.yml) operations file to make the additional changes
required to turn on LDAP authentication.  You will need to provide the following variables in a variables file or with the `-v` option on the BOSH command line:
- `nfs-ldap-service-user`: ldap service account user name
- `nfs-ldap-service-password`: ldap service account password
- `nfs-ldap-host`: ldap server host name or ip address
- `nfs-ldap-port`: ldap server port
- `nfs-ldap-proto`: ldap server protocol (tcp or udp)
- `nfs-ldap-fqdn`: ldap fqdn for user records we will search against when looking up user uids

## Testing

If you want to test against a reference LDAP implementation rather than connecting to your own LDAP server, then you can deploy a sample server by building and uploading this 
[openldap_boshrelease](https://github.com/EMC-Dojo/openldap-boshrelease) release.  This is a fork from a release in `cloudfoundry-community` that also sets up some test 
accounts for you, and installs schema to mirror default user records found in ADFS.

Once you have built and uploaded this bosh release, you can add the LDAP server VM by including the following operations file in your Cloud Foundry
deployment:
[https://github.com/cloudfoundry/persi-ci/blob/master/operations/use-openldap-release.yml](https://github.com/cloudfoundry/persi-ci/blob/master/operations/use-openldap-release.yml)

If you're using this test server, you can use these variable values to connect to it:
- `nfs-ldap-service-user`: cn=admin,dc=domain,dc=com
- `nfs-ldap-service-password`: secret
- `nfs-ldap-host`: openldap.service.cf.internal 
- `nfs-ldap-port`: 389
- `nfs-ldap-proto`: tcp
- `nfs-ldap-fqdn`: ou=Users,dc=domain,dc=com


Once you have redeployed your broker and driver, steps to test will be more or less the same as before, except that when you bind your 
application to your volume, you must specify `username` and `password` instead of UID and GID.  Accordingly, to use the test LDAP server with the nfstestserver, your 
create-service and bind-service commands should look something like this:

```bash
$ cf create-service nfs Existing myVolume -c '{"share":"nfstestserver.service.cf.internal/export/users"}'
$ cf bind-service pora myVolume -c '{"username":"user1000","password":"secret"}'
```

## Credential Rotation

Note that user credentials will be stored as part of the service binding and checked whenever an application is placed on a cell.  Hence, if the password changes on your
LDAP server, the  application must be re-bound to the service and restaged, otherwise it will fail the next time the application is restarted or scaled.
