package internal

import (
	"archive/zip"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	"log"
)

type SubImager interface {
	image.Image
	SubImage(r image.Rectangle) image.Image
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
func CutImage(img image.Image, pieceWidth int, pieceHeigth int) [][]image.Image {
	bankDx := img.Bounds().Dx()
	bankDy := img.Bounds().Dy()
	subImager, err := castSubImager(img)
	if err != nil {
		log.Fatalf("error on sub imager casting: %v", err)
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
	return images
}

func PackImages(dest *zip.Writer, images [][]image.Image, namePrefix string) error {
	if namePrefix != "" {
		namePrefix = namePrefix + "_"
	}
	singsX := len(images)/10 + 1
	singsY := len(images[0])/10 + 1
	fileNameTemplate := fmt.Sprintf("%%s%%0%ddx%%0%dd.jpeg", singsX, singsY)
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

// originalImageFileReader, err := os.Open("file.jpg")
// if err != nil {
// 	log.Fatalf("error on open file: %v", err)
// }

// defer originalImageFileReader.Close()
// i, format, err := image.Decode(originalImageFileReader)
// if err != nil {
// 	log.Fatalf("error on decode file: %v", err)
// }
// log.Printf("Decoded format is: %s", format)

// archive, err := os.Create("archive.zip")
// if err != nil {
// 	log.Fatalf("error on create archive file: %v", err)
// }

// //Create zip Writer
// zipWriter := zip.NewWriter(archive)
// defer zipWriter.Close()

// images := internal.CutImage(i, 100, 100)
// err = internal.PackImages(zipWriter, images, "puzzle")
// if err != nil {
// 	log.Fatalf("error on create archive file: %v", err)
// }
