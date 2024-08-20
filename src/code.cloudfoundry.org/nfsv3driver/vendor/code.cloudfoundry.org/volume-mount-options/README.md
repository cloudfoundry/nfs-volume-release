# volume-mount-options

This is a library that is imported by various parts of CloudFoundry volume services.

## Historical tags

In the past, `v1.0.0` and `v1.1.0` tags were created, and then versioning reverted back to `v0.3.0`.
This caused problems when updating, as the Go toolchain would often try a numeical upgrade which
was actually a chronological downgrade.

As a result, the `v1.0.0` and `v1.1.0` had [retract directives](https://go.dev/ref/mod#go-mod-file-retract)
added, so builds that rely on these versions will still work, but you should not get an update to them.

