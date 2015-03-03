azb
===

A self-contained command-line tool that provides access to Azure Blob storage

## Usage

```
Usage:
  azb [ -F configFile ] [ -e environment ] [ --json ] ls [ <blobspec> ] 
  azb [ -F configFile ] [ -e environment ] [ --json ] tree <container>
  azb [ -F configFile ] [ -e environment ] [ --json ] get <blobpath> [ <dst> ]
  azb [ -F configFile ] [ -e environment ] [ --json ] put <blobpath> [ <src> ]
  azb [ -F configFile ] [ -e environment ] [ --json ] rm [ -f ] <blobpath>
  azb -h | --help
  azb --version

Arguments:
  container     	The name of the container to query
  blobspec      	A reference to one or more blobs (e.g. "mycontainer/foo", 
"mycontainer/")
  blobpath			The path of a blob (e.g. "mycontainer/foo.txt")

Options:
  -e environment    Specifies the Azure Storage Services account to use [default: default]
  -F configFile  	Specifies an alternative per-user configuration file [default: 
/etc/azb/config]
  -f                Forces a destructive operation
  -h, --help     	Show this screen.
  --version     	Show version.

The most commonly used commands are:
   ls         	Lists containers and blobs
   get          Downloads a blob
   put          Uploads a blob
   tree         Prints the contents of a container as a tree
   rm           Deletes a blob
```

## Configuration

`azb` looks for a configuration file containing your Azure Storage credentials.

```TOML
[default]
subscription_id = "SUBSCRIPTION_ID"
storage_account_name = "STORAGE_ACCOUNT_NAME"
storage_account_access_key = "YOU_GET_THE_IDEA"
management_certificate = "JUST_SET_THIS_TO_EMPTY_FOR_NOW"
```

By default, `azb` uses the `default` environment in `/etc/azb/config`, but this can be 
overridden with `-e` and `-F`, respectively.

## Building

`azb` is built using Go 1.3.1 and Goop (`go get github.com/nitrous-io/goop`).

```Bash
$ git clone https://github.com/politician/azb.go.git
$ make install
$ make
$ make test
```

**Note:** There is pretty much no chance that `go get` will work on this repo, and that's _just fine_ with me.

## Packaging

`azb` is a tool that doesn't need (or want) to be installed into the Go bin path.  Instead, we wrap it up into a tarball.

```Bash
$ make archive
$ ls tmp/
```

## Cross-compilation

It should be possible to cross-compile `azb` using a tool like `goxc`, but I haven't tested it on Windows, so there might be some folder path issues (of the / vs \ variety).
