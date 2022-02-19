package img2asci

import (
	_ "image/gif" // https://pkg.go.dev/image#pkg-overview
	_ "image/jpeg"
	_ "image/png"

	"bufio"
	"image"
	"image/color"
	"io"
	"log"
	"os"

	"github.com/disintegration/imaging"
)

const (
	DeafultGrayScale = ".:-=+*#%@" // http://paulbourke.net/dataformats/asciiart/

	DeafultSharp    = 5.0
	DeafultBright   = 5.0
	DeafultContrast = 75.0
)

type Config struct {
	Log              *log.Logger
	ProcessingValues *ProcessingValues
	Term             bool
}

func (c *Config) defaults() {
	if c.Log == nil {
		c.Log = log.Default()
	}

	if c.ProcessingValues == nil {
		c.ProcessingValues = &ProcessingValues{}
		c.ProcessingValues.defaults()
	}
}

type ProcessingValues struct {
	Width  int
	Height int

	Sharp    float64
	Bright   float64
	Contrast float64

	GrayScaleAsciTable string
}

func (pv *ProcessingValues) defaults() {
	if pv.Width == 0 {
		pv.Width = 240
	}

	if pv.Sharp == 0 {
		pv.Sharp = DeafultSharp
	}

	if pv.Bright == 0 {
		pv.Bright = DeafultBright
	}

	if pv.Contrast == 0 {
		pv.Contrast = DeafultContrast
	}

	if pv.GrayScaleAsciTable == "" {
		pv.GrayScaleAsciTable = DeafultGrayScale
	}
}

func (c *Config) Process(inputPath string, outputPath string) error {
	c.defaults()

	img, err := c.loadImage(inputPath)
	if err != nil {
		return err
	}

	buffer, err := c.createBuffer(outputPath, c.Term)
	if err != nil {
		return err
	}
	defer buffer.Flush()

	buf, err := c.run(img, buffer)
	if err != nil {
		return err
	}

	buf.Flush()

	return nil
}

func (c *Config) loadImage(path string) (image.Image, error) {
	c.Log.Println("Loading image:", path)

	imgFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer imgFile.Close()

	c.Log.Println("Decoding image...")
	imgDecoded, imgFormat, err := image.Decode(imgFile)
	if err != nil {
		return nil, err
	}
	c.Log.Println("Image format:", imgFormat)

	return imgDecoded, nil
}

func (c *Config) createBuffer(outputFilePath string, term bool) (*bufio.Writer, error) {
	c.Log.Println("Creating output file:", outputFilePath)

	outputFile, err := os.OpenFile(outputFilePath, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0666)
	if err != nil {
		return nil, err
	}

	if term {
		return bufio.NewWriter(io.MultiWriter(outputFile, os.Stdout)), nil
	}

	return bufio.NewWriter(io.MultiWriter(outputFile)), nil
}

func (c *Config) preprocess(img image.Image) image.Image {
	c.Log.Println("Preprocessing image...")
	if c.ProcessingValues.Height == 0 {
		c.ProcessingValues.Height = (img.Bounds().Max.X * c.ProcessingValues.Width * 10) / (img.Bounds().Max.X * 16)
	}

	img = imaging.Resize(img, c.ProcessingValues.Width, c.ProcessingValues.Height, imaging.Lanczos)
	img = imaging.Sharpen(img, c.ProcessingValues.Sharp)
	img = imaging.AdjustBrightness(img, c.ProcessingValues.Bright)
	img = imaging.AdjustContrast(img, c.ProcessingValues.Contrast)

	return img
}

func (c *Config) run(img image.Image, buffer *bufio.Writer) (*bufio.Writer, error) {
	c.Log.Println("Processing image...")

	img = c.preprocess(img)
	imgX := img.Bounds().Max.X
	imgY := img.Bounds().Max.Y

	for y := 0; y < imgY; y++ {
		for x := 0; x < imgX; x++ {
			gray := color.GrayModel.Convert(img.At(x, y))
			r, g, b, _ := gray.RGBA()
			gr := (19595*r + 38470*g + 7471*b + 1<<15) >> 24

			grayScaled := (gr * (uint32(len(c.ProcessingValues.GrayScaleAsciTable)) - 1)) / 255 //cast uint32 to int can be problematic
			if _, err := buffer.WriteString(string(c.ProcessingValues.GrayScaleAsciTable[grayScaled])); err != nil {
				return nil, err
			}
		}

		if err := buffer.WriteByte('\n'); err != nil {
			return nil, err
		}
	}

	return buffer, nil
}
