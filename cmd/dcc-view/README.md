## Usage
```
$ dcc-view -dcc $DCC_FILE_PATH
```

A DCC file always requires a 256 color palette to be viewed,
since the DCC image format does not include a palette in the
image data, the default palette is greyscale. This default 
greyscale palette contains colors such that the index is the
R, G,and B values for the color.

In order to display assets using a palette, a GPL palette
file must be created. A GPL palette file can be created 
in image editing software, or with the [PL2 transcoder][pl2], 
or with a [GPL transcoder][gpl].

In the case of using a PL2, the palette can be extracted to generate 
a GPL file using the `pl2-to-gpl` command-line tool.
```
dcc-view -dcc $DCC_FILE_PATH -pal $GPL_FILE_PATH
```

[![Product Name Screen Shot][product-screenshot]](#)

[product-screenshot]: ../../assets/dcc_viewer.webp
[pl2]: https://github.com/gravestench/pl2
[gpl]: https://github.com/gravestench/gpl