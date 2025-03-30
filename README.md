# reximage

## A Go package for importing/exporting REXPaint's .xp images

[REXPaint](https://www.gridsagegames.com/rexpaint) is a fabulous program for creating ASCII art, made by ultra-famous roguelike developer Kyzrati. This package allows you to import and export image data to/from the .xp file format produced by REXPaint for use in your project. It was made as part of the larger [Tyumi](https://www.github.com/bennicholls/tyumi) engine, but doesn't depend on it at all and is fine to be used standalone!

## Documentation

Obtain this package in the usual way:

`go get github.com/bennicholls/reximage`

### Importing

```Go
image, err := reximage.Import(pathname)
```

ensuring that `pathname` is a string describing a path to an .xp file. It will return an error if the pathname is invalid or the file cannot be read for some other reason.

Access the cell data using:

```Go
cell, err := image.GetCell(x, y)
```

where x and y are coordinates to a cell in the image (bounded by `image.Width, image.Height`). It will throw an error if (x,y) is not in bounds or if the image hasn't been imported yet. My engine `Tyumi` was originally built around SDL, so `reximage` follows SDL's convention of setting the (0,0) coordinate in the top-left of the image.

`cell` consists of a `Glyph` (ASCII codepoint) and RGB components for both foreground and background colours. Each colour component is 8bits (0-255). You can extract 32bit colours using some helper functions:

```Go
foregroundRGBA, backgroundRGBA := cell.RGBA()
foregroundARGB, backgroundARGB := cell.ARGB()
```

I imagine these are the most popular colour formats but I don't have the research to back that up. Of course, if your program uses a different colour format you can access the individual components and form the colour yourself.

### Exporting

You can also create your own image and export it to the .xp format. First initialize an image:

```Go
var image reximage.ImageData
image.Init(width, height)
```

The cells in the image are initialized to REXPaint's default cell state. Then create cells with the visuals you want and set them with image.SetCell. Some helper functions are provided for setting cell colours with different colour formats.

```Go
var cell reximage.CellData
cell.Glyph = 88 // An 'X'
cell.SetColoursARGB(0xFFFFFFFF, 0xFF000000) // foreground white, background black
image.SetCell(0, 0, cell) // (0, 0) is the (x, y) coordinate of the top left cell in the image
```

If you use a different colour format than ARGB or RGBA then you'll have to set the cell's R,G,B components yourself. Once your image has been filled with cell data you can export it to a file:

```Go
err := reximage.Export(image, "some_file_path.xp")
```

If the file already exists it will be overwritten. If the export process fails Export will return an error.

Complete(-ish) documentation can be found on [GoDoc.org](https://godoc.org/github.com/bennicholls/reximage).

## Limitations

REXPaint supports saving images with up to 9 layers, but at the moment reximage does not retain layer data in xp files. On import, images with multiple layers are painted bottom to top and the final composed image is returned. Exported images are limited to 1 layer, encoding multiple layers is not yet implemented.

## Future

This is all my engine needs at the moment so I'm not sure what else would be helpful here. On the off-chance someone else is using this, please let me know if there's anything you think it could use.

## License

reximage is licensed under the MIT license (see LICENSE file).
