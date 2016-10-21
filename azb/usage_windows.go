package main

func init() {
	usageMsg = `azb - an uncomplicated azure blob storage client

Usage:
  azb [ -F configFile ] [ -e environment ] [ --json ] ls [ <blobspec> ]
  azb [ -F configFile ] [ -e environment ] [ --json ] tree <container>
  azb [ -F configFile ] [ -e environment ] [ --json ] get <blobpath> [ <dst> ]
  azb [ -F configFile ] [ -e environment ] [ --json ] put <blobpath> [ <src> ]
  azb [ -F configFile ] [ -e environment ] [ --json ] rm [ -f ] <blobpath>
  azb [ -F configFile ] [ -e environment ] [ --json ] [ -w workers ] size [ - | <blobspecs>... ]
  azb -h | --help
  azb --version

Arguments:
  container      The name of the container to query.
  blobspec       A reference to one or more blobs (e.g. "mycontainer/foo", "mycontainer/")
  blobpath       The path of a blob (e.g. "mycontainer/foo.txt")

Options:
  -e environment  Specifies the Azure Storage Services account to use [default: default]
  -F configFile   Specifies an alternative per-user configuration file [default: C:\_azb.toml]
  -f              Forces a destructive operation
  -w workers      The maximum number of concurrent workers to use [default: 10]
  -h, --help      Show this screen.
  --version       Show version.
`
}
