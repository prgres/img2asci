package main

import (
	"bufio"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/prgres/img2asci/pkg/img2asci"
	"github.com/urfave/cli/v2"
)

func createBuffer(outputFilePath string, term bool) (*bufio.Writer, error) {
	log.Println("Creating output file:", outputFilePath)
	outputFile, err := os.OpenFile(outputFilePath, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0666)
	if err != nil {
		return nil, err
	}

	if term {
		return bufio.NewWriter(io.MultiWriter(outputFile, os.Stdout)), nil
	}

	return bufio.NewWriter(io.MultiWriter(outputFile)), nil
}

func process(sourceFilePath, outputFilePath string, pv *img2asci.ProcessingValues, term bool) error {
	img, err := img2asci.LoadImage(sourceFilePath)
	if err != nil {
		return err
	}

	buffer, err := createBuffer(outputFilePath, term)
	if err != nil {
		return err
	}
	defer buffer.Flush()

	buf, err := pv.Process(img, buffer)
	if err != nil {
		return err
	}

	buf.Flush()
	return nil
}

func main() {
	flags := []cli.Flag{
		&cli.IntFlag{
			Name:    "width",
			Aliases: []string{"wx"},
			Value:   80,
			Usage:   "width of output image ",
		},
		&cli.IntFlag{
			Name:    "height",
			Aliases: []string{"hx"},
			Value:   0,
			Usage:   "width of output image",
		},
		&cli.Float64Flag{
			Name:    "sharp",
			Aliases: []string{"vs"},
			Value:   img2asci.DeafultSharp,
			Usage:   "preprocessing sharpening value",
		},
		&cli.Float64Flag{
			Name:    "bright",
			Aliases: []string{"vb"},
			Value:   img2asci.DeafultBright,
			Usage:   "preprocessing bright value",
		},
		&cli.Float64Flag{
			Name:    "contrast",
			Aliases: []string{"vc"},
			Value:   img2asci.DeafultContrast,
			Usage:   "preprocessing contrast value",
		},
		&cli.StringFlag{
			Name:    "outputFilePath",
			Aliases: []string{"o"},
			Value:   "./output.txt",
			Usage:   "name of output file",
		},
		&cli.StringFlag{
			Name:  "grayScaleAsciTable",
			Value: img2asci.DeafultGrayScale,
			Usage: "override gray scale with ASCI characters",
		},
		&cli.BoolFlag{
			Name:    "term",
			Aliases: []string{"t"},
			Value:   false,
			Usage:   "if print to term",
		},
	}

	app := &cli.App{
		EnableBashCompletion: true,
		Name:                 "img2asci",
		Usage:                "ASCI ART converter",
		Description:          "Simple but yet very powerfull converter from ordinary, plain images to beatiful ASCI art",
		Flags:                flags,
		Compiled:             time.Now(),
		Authors: []*cli.Author{
			{
				Name: "M. Więcek",
			},
		},
		Action: func(c *cli.Context) error {
			if c.NArg() == 0 {
				return cli.ShowAppHelp(c)
			}

			return process(
				strings.TrimSpace(c.Args().First()),
				c.String("outputFilePath"),
				&img2asci.ProcessingValues{
					Width:              c.Int("width"),
					Height:             c.Int("height"),
					Sharp:              c.Float64("sharp"),
					Bright:             c.Float64("bright"),
					Contrast:           c.Float64("contrast"),
					GrayScaleAsciTable: c.String("grayScaleAsciTable"),
				},
				c.Bool("term"),
			)
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
