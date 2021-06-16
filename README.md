
# goup

This is a quick Go program to install or upgrade your Go runtime by downloading the latest release from golang.org:

    % sudo goup
    You are running Go 1.10.1 for linux (amd64)
    Go 1.10.3 linux amd64 is available in tar format
    Download and install? y
    Downloading go1.10.3.linux-amd64.tar.gz
    Fetching https://dl.google.com/go/go1.10.3.linux-amd64.tar.gz
    Downloaded and SHA256 checked
    Go upgraded successfully
    % go version
    go version go1.10.3 linux/amd64
    %

It should work for macOS and Linux, patches for Windows compatibility would be welcomed.

## Features

 - Interactive
 
 - Examines versions to see if there's an upgrade

 - Checks SHA256 checksums
 
 - Checks byte count of download

 - Not a shell script and doesn't encourage you to do `curl ... | sh`
 
 - Sets read permissions on unpacked files for use from `sudo`
 
 - Allows you to override arch or OS (see `--help`)
 
 - Usable for initial install, if you [download  a built binary](https://github.com/lpar/goup/releases)

 - Doesn't attempt to manage having multiple different versions of Go
   installed, because that's already supported by the standard `go get` command
   once you have the latest version installed. See [ the Go
   documentation](https://golang.org/doc/install#extra_versions) for details.
 
## Bugs and limitations

 - Slurps the entire downloaded distribution file into RAM.

 - Doesn't work on Windows yet.

 - If the Go team distributes a tar file that doesn't have everything under a
   top level "go" directory, it'll fail (and leave junk in your destination
   directory).

 - Doesn't handle multiple versions. If you want to use multiple versions of Go, install the latest one, then use Go's regular features to handle it:

    go get golang.org/dl/go1.16.1
    go1.16.1 build

