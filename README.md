
# goup

This is a quick Go program to upgrade your Go runtime by downloading the latest
release from golang.org:

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

Currently it assumes Go is installed from tar file in `/usr/local/go`, so it
should work for macOS and Linux. If you keep Go elsewhere for some reason you
can change the constant in the source code.

## Features

 - Interactive

 - Examines versions to see if there's an upgrade

 - Checks SHA256 checksums

## Bugs and limitations

 - Slurps the entire downloaded distribution file into RAM.

 - Doesn't work on Windows.

 - If the Go team distributes a tar file that doesn't have everything under a
   top level "go" directory, it'll fail (and leave junk in your destination
   directory).

