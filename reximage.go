// A package for decoding and handling .xp files produced by Kyzrati's fabulous REXPaint program, the gold-standard in
// ASCII art drawing programs. It can be found at www.gridsagegames.com/rexpaint.
//
// reximage is part of the Tyumi engine by Benjamin Nicholls, but feel free to use it as a standalone package!
package reximage

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"errors"
	"os"
	"strings"
)

// ImageData is the struct holding the decoded and exported image data.
type ImageData struct {
	Width  int
	Height int
	Cells  []CellData //will have Width*Height Elements
}

// Init initializes an image with the provided size, setting all cells to the REXPaint default.
func (id *ImageData) Init(width, height int) {
	id.Width = width
	id.Height = height
	id.Cells = make([]CellData, width*height)
	for i := range id.Cells {
		id.Cells[i].Clear()
	}
}

// GetCell returns the CellData at coordinate (x, y) of the decoded image, with (0,0) at the top-left of the image.
func (id ImageData) GetCell(x, y int) (cd CellData, err error) {
	if len(id.Cells) == 0 {
		err = errors.New("Image has no data.")
		return
	}

	if x >= id.Width || y >= id.Height || x < 0 || y < 0 {
		err = errors.New("x, y coordinates out of bounds.")
		return
	}

	cd = id.Cells[x+y*id.Width]

	return
}

// SetCell sets the cell at (x, y) to cell. (0, 0) is considered to be the top-left of the image, with positive y
// values moving down.
func (id *ImageData) SetCell(x, y int, cell CellData) (err error) {
	if x >= id.Width || y >= id.Height || x < 0 || y < 0 {
		return errors.New("x, y coordinates out of bounds.")
	}

	id.Cells[x+y*id.Width] = cell

	return nil
}

// CellData holds the decoded data for a single cell. Colours are split into uint8 components so the user can combine
// them into whatever colour format they need. Some popular colour format conversion functions are provided as well.
type CellData struct {
	Glyph uint32 // ASCII code for glyph
	R_f   uint8  // Foreground Colour - Red channel
	G_f   uint8  // Foreground Colour - Green channel
	B_f   uint8  // Foreground Colour - Blue channel
	R_b   uint8  // Background Colour - Red channel
	G_b   uint8  // Background Colour - Green channel
	B_b   uint8  // Background Colour - Blue channel
}

// ARGB returns the foreground and background colours of the cell in ARGB format.
// Alpha in this case is always set to maximum (255).
func (cd CellData) ARGB() (fore, back uint32) {
	fore = uint32(0xFF << 24) //set alpha to 255
	fore |= uint32(cd.R_f) << 16
	fore |= uint32(cd.G_f) << 8
	fore |= uint32(cd.B_f)

	back = uint32(0xFF << 24) //set alpha to 255
	back |= uint32(cd.R_b) << 16
	back |= uint32(cd.G_b) << 8
	back |= uint32(cd.B_b)

	return
}

// SetColoursRGBA sets the colours of the cell, interpreting the input uint32 colours in ARGB8888 format.
// If background alpha is 0, the cell is set to undrawn. Foreground alpha is ignored.
func (cd *CellData) SetColoursARGB(fore, back uint32) {
	if uint8(back>>24) == 0 {
		cd.B_b, cd.G_b, cd.R_b = 0xFF, 0, 0xFF
		return
	}

	cd.B_f = uint8(fore & 0xFF)
	cd.G_f = uint8((fore >> 8) & 0xFF)
	cd.R_f = uint8((fore >> 16) & 0xFF)

	cd.B_b = uint8(back & 0xFF)
	cd.G_b = uint8((back >> 8) & 0xFF)
	cd.R_b = uint8((back >> 16) & 0xFF)
}

// RGBA returns the foreground and background colours of the cell in RGBA format.
// Alpha in this case is always set to maximum (255).
func (cd CellData) RGBA() (fore, back uint32) {
	fore = uint32(cd.R_f) << 24
	fore |= uint32(cd.G_f) << 16
	fore |= uint32(cd.B_f) << 8
	fore |= 0xFF //set alpha to 255

	back = uint32(cd.R_b) << 24
	back |= uint32(cd.G_b) << 16
	back |= uint32(cd.B_b) << 8
	back |= 0xFF //set alpha to 255

	return
}

// SetColoursRGBA sets the colours of the cell, interpreting the input uint32 colours in RGBA8888 format.
// If background alpha is 0, the cell is set to undrawn. Foreground alpha is ignored.
func (cd *CellData) SetColoursRGBA(fore, back uint32) {
	if uint8(back&0xFF) == 0 {
		cd.R_b, cd.G_b, cd.B_b = 0xFF, 0, 0xFF
		return
	}

	cd.B_f = uint8((fore >> 8) & 0xFF)
	cd.G_f = uint8((fore >> 16) & 0xFF)
	cd.R_f = uint8((fore >> 24) & 0xFF)

	cd.B_b = uint8((back >> 8) & 0xFF)
	cd.G_b = uint8((back >> 16) & 0xFF)
	cd.R_b = uint8((back >> 24) & 0xFF)
}

// Undrawn returns whether the cell is "undrawn" or empty, which in XP files is identified by the background
// colour (255, 0, 255).
func (cd CellData) Undrawn() bool {
	return cd.R_b == 255 && cd.G_b == 0 && cd.B_b == 255
}

// Clear removes the cell's glyph and sets the cell's colours to the rexpaint default. A cleared cell will be considered
// undrawn.
func (cd *CellData) Clear() {
	cd.Glyph = 0
	cd.R_b, cd.G_b, cd.B_b = 0xFF, 0, 0xFF
	cd.R_f, cd.G_f, cd.B_f = 0, 0, 0
}

// Import imports an image from the xp file at the provided path. Returns the Imagedata and an error. If an error is
// present, ImageData will be no good.
func Import(path string) (image ImageData, err error) {
	image = ImageData{}

	if !strings.HasSuffix(path, ".xp") {
		err = errors.New("File is not an XP image.")
		return
	}

	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	//xp data is gzipped
	data, err := gzip.NewReader(f)
	if err != nil {
		return
	}
	defer data.Close()

	//read rexpaint version num and the number of layers
	var version int32
	var numLayers uint32
	err = binary.Read(data, binary.LittleEndian, &version)
	err = binary.Read(data, binary.LittleEndian, &numLayers)
	if err != nil {
		return
	}

	//read into the first layer so we can get the image dimensions and initialize cell data
	var w, h uint32
	err = binary.Read(data, binary.LittleEndian, &w)
	err = binary.Read(data, binary.LittleEndian, &h)
	if err != nil {
		return
	}

	image.Init(int(w), int(h))

	//read layers, painting from lowest layer to highest
	for layer := range int(numLayers) {
		if layer != 0 {
			//if reading subsequent layers, throw away the dimension bytes since we've already read them before
			err = binary.Read(data, binary.LittleEndian, &w)
			err = binary.Read(data, binary.LittleEndian, &h)
		}

		for i := range image.Width * image.Height {
			//read bytes for each cell.
			c := CellData{}
			err = binary.Read(data, binary.LittleEndian, &c)
			if err != nil {
				return
			}

			//xp images are encoded in the totally insane column-major order for some reason, we correct that here
			//(sorry Kyzrati, gotta put my foot down on this one)
			image.SetCell(i/image.Height, i%image.Height, c)
		}
	}

	return
}

// Export encodes an image as an .xp file and writes to disk at the specified path. If a file already exists at that
// location it is overwritten.
func Export(image ImageData, path string) (err error) {
	if !strings.HasSuffix(path, ".xp") {
		path += ".xp"
	}

	f, err := os.Create(path)
	if err != nil {
		return
	}
	defer f.Close()

	imagebuffer := new(bytes.Buffer)
	binary.Write(imagebuffer, binary.LittleEndian, int32(-1)) // version number
	binary.Write(imagebuffer, binary.LittleEndian, uint32(1)) // number of layers
	binary.Write(imagebuffer, binary.LittleEndian, uint32(image.Width))
	binary.Write(imagebuffer, binary.LittleEndian, uint32(image.Height))

	for x := range image.Width {
		for y := range image.Height {
			cell, _ := image.GetCell(x, y)
			binary.Write(imagebuffer, binary.LittleEndian, cell)
		}
	}

	zipper := gzip.NewWriter(f)
	defer zipper.Close()

	_, err = zipper.Write(imagebuffer.Bytes())

	return
}
