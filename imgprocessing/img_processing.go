package imgprocessing

import (
	"archive/zip"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png" // registers img formats
	"log"
	"os"
)

type subImager interface {
	image.Image
	SubImage(r image.Rectangle) image.Image
}

// func ProcessImage(fileName string, dx int, dy int) (string, error) {
// 	// открываем изображение
// 	// декодируем формат
// 	img, format, err := OpenImage(fileName)

// 	log.Printf("Decoded format is: %s", format)

// 	// создаём архив
// 	name := strings.TrimSuffix(origFile.Name(), filepath.Ext(origFile.Name()))
// 	archive, err := os.Create(fmt.Sprintf("%s.zip", name))
// 	if err != nil {
// 		log.Printf("error on create archive file: %v", err)
// 		return "", err
// 	}

// 	zipWriter := zip.NewWriter(archive)
// 	defer zipWriter.Close()

// 	// режем изображение
// 	images, err := cutImage(img, dx, dy)
// 	if err != nil {
// 		log.Printf("error on cut img: %v", err)
// 		return "", err
// 	}

// 	// пакуем в архив
// 	if err = packImages(zipWriter, images, filepath.Base(name)); err != nil {
// 		log.Printf("error on create archive file: %v", err)
// 		return "", err
// 	}
// 	return archive.Name(), nil
// }

func OpenImage(fileName string) (img image.Image, imgFormat string, err error) {
	origFile, err := os.Open(fileName)
	if err != nil {
		log.Printf("error opening file: %v", err)
		return nil, "", err
	}

	img, imgFormat, err = image.Decode(origFile)
	if err != nil {
		log.Printf("error on decode file: %v", err)
		return nil, "", err
	}

	return img, imgFormat, nil
}

func castSubImager(img image.Image) (subImager, error) {
	switch image := img.(type) {
	case *image.Alpha:
		log.Printf("%T", image)
		return image, nil
	case *image.Alpha16:
		log.Printf("%T", image)
		return image, nil
	case *image.CMYK:
		log.Printf("%T", image)
		return image, nil
	case *image.Gray:
		log.Printf("%T", image)
		return image, nil
	case *image.Gray16:
		log.Printf("%T", image)
		return image, nil
	case *image.NRGBA:
		log.Printf("%T", image)
		return image, nil
	case *image.NRGBA64:
		log.Printf("%T", image)
		return image, nil
	case *image.NYCbCrA:
		log.Printf("%T", image)
		return image, nil
	case *image.Paletted:
		log.Printf("%T", image)
		return image, nil
	case *image.RGBA:
		log.Printf("%T", image)
		return image, nil
	case *image.RGBA64:
		log.Printf("%T", image)
		return image, nil
	case *image.YCbCr:
		log.Printf("%T", image)
		return image, nil
	}

	return image.NewAlpha(img.Bounds()), errors.New("unknown format")
}

// note: every unit of [][]image.Image shares pixels with img
func CutImage(img image.Image, pieceWidth int, pieceHeigth int) ([][]image.Image, error) {
	if pieceWidth < 32 || pieceHeigth < 32 {
		return nil, errors.New("cut too small")
	}

	bankDx := img.Bounds().Dx()
	bankDy := img.Bounds().Dy()

	subImager, err := castSubImager(img)
	if err != nil {
		log.Printf("error on sub imager casting: %v", err)
		return nil, err
	}

	log.Printf("dimension banks x: %d, y: %d", bankDx, bankDy)

	images := make([][]image.Image, 0)

	for y := 0; y*pieceHeigth < bankDy; y++ {
		s := make([]image.Image, 0)
		images = append(images, s)
		images[y] = make([]image.Image, 0)

		for x := 0; x*pieceWidth < bankDx; x++ {
			anotherImage := subImager.SubImage(image.Rect(x*pieceWidth, y*pieceHeigth, (x+1)*pieceWidth, (y+1)*pieceHeigth))
			images[y] = append(images[y], anotherImage)
		}
	}

	return images, nil
}

func PackImages(dest *zip.Writer, images [][]image.Image, namePrefix string) error {
	if namePrefix != "" {
		namePrefix += "_"
	}

	digitsByX := countDigits(len(images))
	digitsByY := countDigits(len(images[0]))

	// "%s%0Xdx%0Yd.jpeg", digitsByX = 2, digitsByY = 4 -> "%s%02dx%04d.jpeg"
	fileNameTemplate := fmt.Sprintf("%%s%%0%ddx%%0%dd.jpeg", digitsByX, digitsByY)
	options := jpeg.Options{Quality: 100}

	for x, sliceByX := range images {
		for y, image := range sliceByX {
			w, err := dest.Create(fmt.Sprintf(fileNameTemplate, namePrefix, x+1, y+1))
			if err != nil {
				return err
			}

			if err := jpeg.Encode(w, image, &options); err != nil {
				return err
			}
		}
	}

	return nil
}

func countDigits(i int) int {
	count := 0

	for i > 0 {
		i /= 10
		count++
	}

	return count
}