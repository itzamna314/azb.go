azb
===

A golang command-line tool that provides familiar access to Azure Blob storage

## Usage

```
Usage:
  azb [ -F configFile ] [ -e environment ] [ --json ] ls [ <blobspec> ] 
  azb [ -F configFile ] [ -e environment ] [ --json ] pull <blobpath> [ <dst> ]
  azb [ -F configFile ] [ -e environment ] [ --json ] push [ -R ] <blobpath> [ <src> ]
  azb [ -F configFile ] [ -e environment ] [ --json ] rm <blobpath>
  azb [ -F configFile ] [ -e environment ] [ --json ] cp <srcblobpath> <dstblobpath>
  azb [ -F configFile ] [ -e environment ] [ --json ] mv <srcblobpath> <dstblobpath>
  azb -h | --help
  azb --version

Arguments:
  container     	The name of the container to query
  blobspec      	A reference to one or more blobs (e.g. "mycontainer/foo*", 
"mycontainer/")
  blobpath			The path of a blob (e.g. "mycontainer/foo.txt")

Options:
  --add=NAME    	Creates a container
  -e environment    Specifies the Azure Storage Services account to use [default: default]
  -F configFile  	Specifies an alternative per-user configuration file [default: 
/etc/azb/config]
  -h, --help     	Show this screen.
  --list        	List all of the containers in the environment
  --rm=NAME     	Destroys a container
  --version     	Show version.

The most commonly used azb commands are:
   ls         	Lists containers and blobs
   pull         Downloads blobs
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

