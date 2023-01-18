package internal

import (
	"archive/zip"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type SubImager interface {
	image.Image
	SubImage(r image.Rectangle) image.Image
}

func ProcessImage(fileName string, dx int, dy int) (string, error) {
	origFile, err := os.Open(fileName)
	if err != nil {
		log.Printf("error opening file: %v", err)
		return "", err
	}

	img, format, err := image.Decode(origFile)
	if err != nil {
		log.Printf("error on decode file: %v", err)
		return "", err
	}
	log.Printf("Decoded format is: %s", format)

	name := strings.TrimSuffix(origFile.Name(), filepath.Ext(origFile.Name()))

	archive, err := os.Create(fmt.Sprintf("%s.zip", name))
	if err != nil {
		log.Printf("error on create archive file: %v", err)
		return "", err
	}

	zipWriter := zip.NewWriter(archive)
	defer zipWriter.Close()

	images, err := cutImage(img, dx, dy)
	if err != nil {
		log.Printf("error on cut img: %v", err)
		return "", err
	}
	if err = packImages(zipWriter, images, origFile.Name()); err != nil {
		log.Printf("error on create archive file: %v", err)
		return "", err
	}
	return archive.Name(), nil
}

func castSubImager(img image.Image) (SubImager, error) {
	switch t := img.(type) {
	case *image.Alpha:
		log.Printf("%T", t)
		return t, nil
	case *image.Alpha16:
		log.Printf("%T", t)
		return t, nil
	case *image.CMYK:
		log.Printf("%T", t)
		return t, nil
	case *image.Gray:
		log.Printf("%T", t)
		return t, nil
	case *image.Gray16:
		log.Printf("%T", t)
		return t, nil
	case *image.NRGBA:
		log.Printf("%T", t)
		return t, nil
	case *image.NRGBA64:
		log.Printf("%T", t)
		return t, nil
	case *image.NYCbCrA:
		log.Printf("%T", t)
		return t, nil
	case *image.Paletted:
		log.Printf("%T", t)
		return t, nil
	case *image.RGBA:
		log.Printf("%T", t)
		return t, nil
	case *image.RGBA64:
		log.Printf("%T", t)
		return t, nil
	case *image.YCbCr:
		log.Printf("%T", t)
		return t, nil
	}
	return image.NewAlpha(img.Bounds()), errors.New("unknown format")
}

// note: every unit of [][]image.Image shares pixels with img
func cutImage(img image.Image, pieceWidth int, pieceHeigth int) ([][]image.Image, error) {
	bankDx := img.Bounds().Dx()
	bankDy := img.Bounds().Dy()
	subImager, err := castSubImager(img)
	if err != nil {
		log.Printf("error on sub imager casting: %v", err)
		return nil, err
	}
	log.Printf("dimension banks x: %d, y: %d", bankDx, bankDy)

	images := make([][]image.Image, 0)

	y := 0
	for y*pieceHeigth < bankDy {
		s := make([]image.Image, 0)
		images = append(images, s)
		images[y] = make([]image.Image, 0)

		x := 0
		for x*pieceWidth < bankDx {
			anotherImage := subImager.SubImage(image.Rect(x*pieceWidth, y*pieceHeigth, (x+1)*pieceWidth, (y+1)*pieceHeigth))
			images[y] = append(images[y], anotherImage)
			x++
		}
		y++
	}

	return images, nil
}

func packImages(dest *zip.Writer, images [][]image.Image, namePrefix string) error {
	if namePrefix != "" {
		namePrefix = namePrefix + "_"
	}
	digitsByX := len(images)/10 + 1
	digitsByY := len(images[0])/10 + 1
	fileNameTemplate := fmt.Sprintf("%%s%%0%ddx%%0%dd.jpeg", digitsByX, digitsByY)
	for x, sliceByX := range images {
		for y, image := range sliceByX {
			w, err := dest.Create(fmt.Sprintf(fileNameTemplate, namePrefix, x+1, y+1))
			if err != nil {
				return err
			}
			o := jpeg.Options{
				Quality: 100,
			}
			err = jpeg.Encode(w, image, &o)
			if err != nil {
				return err
			}

		}
	}

	return nil
}
