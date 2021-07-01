<!-- PROJECT LOGO -->
<h1 align="center">DCC</h1>
<p align="center">
  Package for transcoding DCC image files.
  <br />
  <br />
  <a href="https://github.com/gravestench/dcc/issues">Report Bug</a>
  ·
  <a href="https://github.com/gravestench/dcc/issues">Request Feature</a>
</p>

<!-- ABOUT THE PROJECT -->
## About

This package provides a DCC image transcoder implementation.

This package also contains command-line and graphical applications for working with DCC image files.

## Project Structure
* `pkg/` - This directory contains the core DCC transcoder library. This is the directory to import if you want to 
  write new golang applications using this library.
    ```golang
   import (
        dcc "github.com/gravestench/dcc/pkg"
  )
    ```
* `cmd/` - This directory contains command-line and graphical applications, each having their own sub-directory.
* `assets/` - This directory contains files, like the images displayed in this README, or test dcc file data.

## Getting Started

### Prerequisites
You need to install [Go 1.16][golang], as well as set up your go environment. 
In order to install the applications inside of `cmd/`, you will need to 
make sure that `$GOBIN` is defined and points to a valid directory, 
and this will also need to be added to your `$PATH` environment variable.
```shell
export GOBIN=$HOME/.gobin
mkdir -p $GOBIN
PATH=$PATH:$GOBIN
```

### Installation
As long as `$GOBIN` is defined and on your `$PATH`, you can build and install all apps inside of 
`cmd/` by running these commands:

```shell
# clone the repo, enter the dir
git clone http://github.com/gravestench/dcc
cd dcc

# build and install inside of $GOBIN
go build ./cmd/...
go install ./cmd/...
```

At this point, you should be able to run the apps inside of `cmd/` from the command-line, like `dcc-view`.

<!-- CONTRIBUTING -->
## Contributing

I've set up all of the repos with a similar project structure. `~/pkg/` is where the actual
transcoder library is, and `~/cmd/` has subdirectories for each CLI/GUI application that can be
compiled.

Any contributions are **greatly appreciated**.

1. Fork the Project
2. Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3. Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the Branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

<!-- MARKDOWN LINKS & IMAGES -->
[dt1]: https://github.com/gravestench/dt1
[dc6]: https://github.com/gravestench/dc6
[dat_palette]: https://github.com/gravestench/dat_palette
[ds1]: https://github.com/gravestench/ds1
[cof]: https://github.com/gravestench/cof
[golang]: https://golang.org/dl/