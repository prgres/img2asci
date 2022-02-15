package img2asci

import (
	_ "image/gif" // https://pkg.go.dev/image#pkg-overview
	_ "image/jpeg"
	_ "image/png"

	"bufio"
	"image"
	"image/color"
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

type ProcessingValues struct {
	Width  int
	Height int

	Sharp    float64
	Bright   float64
	Contrast float64

	GrayScaleAsciTable string
}

func LoadImage(path string) (image.Image, error) {
	log.Println("Loading image:", path)
	imgFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer imgFile.Close()

	log.Println("Decoding image...")
	imgDecoded, imgFormat, err := image.Decode(imgFile)
	if err != nil {
		return nil, err
	}
	log.Println("Image format:", imgFormat)

	return imgDecoded, nil
}

func (pv *ProcessingValues) preprocessorImage(img image.Image) image.Image {
	log.Println("Preprocessing image...")
	if pv.Height == 0 {
		pv.Height = (img.Bounds().Max.X * pv.Width * 10) / (img.Bounds().Max.X * 16)
	}

	img = imaging.Resize(img, pv.Width, pv.Height, imaging.Lanczos)
	img = imaging.Sharpen(img, pv.Sharp)
	img = imaging.AdjustBrightness(img, pv.Bright)
	img = imaging.AdjustContrast(img, pv.Contrast)

	return img
}

func (pv *ProcessingValues) Process(img image.Image, buffer *bufio.Writer) (*bufio.Writer, error) {
	log.Println("Procesing image...")
	img = pv.preprocessorImage(img)
	imgX := img.Bounds().Max.X
	imgY := img.Bounds().Max.Y

	for y := 0; y < imgY; y++ {
		for x := 0; x < imgX; x++ {
			gray := color.GrayModel.Convert(img.At(x, y))
			r, g, b, _ := gray.RGBA()
			gr := (19595*r + 38470*g + 7471*b + 1<<15) >> 24

			grayScaled := (gr * (uint32(len(pv.GrayScaleAsciTable)) - 1)) / 255 //cast uint32 to int can be problematic
			if _, err := buffer.WriteString(string(pv.GrayScaleAsciTable[grayScaled])); err != nil {
				return nil, err
			}
		}
		if err := buffer.WriteByte('\n'); err != nil {
			return nil, err
		}

	}

	return buffer, nil
}
