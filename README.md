<!-- PROJECT LOGO -->
<h3 align="center">dcc</h3>
<p align="center">
  Individual package focusing on dcc file transcoding
  <br />
  <br />
  <a href="https://github.com/gravestench/dcc/issues">Report Bug</a>
  ·
  <a href="https://github.com/gravestench/dcc/issues">Request Feature</a>
</p>

<!-- ABOUT THE PROJECT -->
## About The Project

[![Product Name Screen Shot][product-screenshot]](https://example.com)

The big-picture idea behind this repo is utilities dealing with dcc file transcoding are 
self-contained. You should not need to install monolithic GUI applications to do simple file 
operations e.g. converting a dcc file into a sequence of PNG files. You should be able to clone the 
source, build and install the CLI/GUI tools and work with these file types easily.

Other repos that deal with transcoding are:
* [pl2]
* [dt1]
* [dc6]
* [dat_palette]
* [ds1]
* [cof]

## Getting Started

### Prerequisites
* [Go 1.16][golang]

### Installation
As long as `$GOBIN` is defined and on your `$PATH`, you can build and install the apps inside of 
`~/cmd` by running these commands:

```
$ git clone http://github.com/gravestench/dcc
$ cd dcc
$ go build ./cmd/...
$ go install ./cmd/...
```
## Usage
After installation is completed, you can run the `dcc-view` app from the repo (located at 
`~/cmd/dcc-view`) by running the following:

```
$ dcc-view -dcc <dcc file>
```

Without a GPL palette file, this will show the colors as grayscale, where the palette index is the
R, G,and B values for the pixel.

To run the dcc viewer with a palette, I had to make the [PL2 transcoder repo][pl2], and then make a
simple app to extract the palette from the pl2 file 
(in this case, `data/global/palette/act1/Pal.pl2` inside of the `d2data.mpq` file). You can pass
any 256 color GPL file to the `dcc-view` app like this:

```
dcc-view -dcc <dcc file> -pal <gpl palette file>
```

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
[product-screenshot]: assets/dcc_viewer.webp
[pl2]: https://github.com/gravestench/pl2
[dt1]: https://github.com/gravestench/dt1
[dc6]: https://github.com/gravestench/dc6
[dat_palette]: https://github.com/gravestench/dat_palette
[ds1]: https://github.com/gravestench/ds1
[cof]: https://github.com/gravestench/cof
[golang]: https://golang.org/dl/