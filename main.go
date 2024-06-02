// PlayStation PlayView Exporter.
//
// See README for more details.
//
// Jörn Lenoch @ 2024

package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"image/draw"
	"image/jpeg"
	"image/png"
	"log"
	"os"
	"strings"
)

var totalDataEntries int
var totalLengthFirstPart int

type PageInfo struct {
	// 0010 	4 	Offset file name.gvd (without header TGDT0100)
	offsetFileName int

	// 0014 	4 	Length file name.gvd (00 is not counted)
	lengthFileName int

	// 0018 	4 	Offset Data Base Viewer
	offsetDataBaseViewer int

	// 001C 	4 	Length Data base Viewer file
	lengthDataBaseViewer int

	fileName string

	imageWidth     int
	imageHeight    int
	lengthDatabase int

	images []ImageInfo

	lengthImages   int
	paramLength    int
	entranceLength int
	imageType      string
}

type ImageInfo struct {
	gridPosW          int
	gridPosH          int
	height            int
	width             int
	fileLength        int
	fileLengthPadding int
	layer             int
}

var pages []PageInfo

func main() {
	gvdHandle, err := os.Open("gvd.dat")
	if err != nil {
		panic(err)
	}

	err = readHeader(gvdHandle)
	if err != nil {
		panic(err)
	}

	err = readFileNames(gvdHandle)
	if err != nil {
		panic(err)
	}

	err = readDatabase(gvdHandle)
	if err != nil {
		log.Panicf("unable to read databases: %v", err)
	}

	log.Print("done")
}

func readDatabase(f *os.File) error {

	for i := int(0); i < totalDataEntries; i++ {

		// Jump to database.
		_, _ = f.Seek(int64(totalLengthFirstPart+pages[i].offsetDataBaseViewer), 0)

		key, _ := readString(f, 16)
		if key == "GVEW0100JPEG0100" {
			pages[i].imageType = "jpeg"
		} else if key == "GVEW0100GVMP0100" {
			pages[i].imageType = "gvmp"
		} else {
			log.Panicf("unknown database type: %v", key)
		}

		// Read Length
		pages[i].imageWidth, _ = readUint32(f)
		log.Printf("[%v] length: %v", i, pages[i].imageWidth)

		// Read Heigth
		pages[i].imageHeight, _ = readUint32(f)
		log.Printf("[%v] height: %v", i, pages[i].imageHeight)

		// Read BLK
		readCompare(f, []byte{0x42, 0x4C, 0x4B, 0x5F})

		// Length Database
		pages[i].lengthDatabase, _ = readUint32(f)
		log.Printf("[%v] lengthDatabase: %v", i, pages[i].lengthDatabase)

		// DATABASES START
		readCompare(f, []byte{00, 00, 00, 01, 00, 00, 00, 00})

		// 0028 	4 	00 00 00 20 	each entrance length: 0X20
		pages[i].entranceLength, _ = readUint32(f)
		// readCompare(f, []byte{0x00, 0x00, 0x00, 0x20})
		log.Printf("[%v] entryLength: %v", i, pages[i].entranceLength)

		// 002C 	4 	00 00 00 04 	each parameter length: 0X04
		pages[i].paramLength, _ = readUint32(f)
		// readCompare(f, []byte{0x00, 0x00, 0x00, 0x04})
		log.Printf("[%v] paramLength: %v", i, pages[i].paramLength)

		// Read images
		numImages := pages[i].lengthDatabase / pages[i].entranceLength
		pages[i].images = make([]ImageInfo, numImages)

		//pos, _ := f.Seek(0, 1)
		//log.Printf(" Starting at %v", pos)

		for j := int(0); j < numImages; j++ {

			if pages[i].paramLength == 4 {
				// 0030 	4 	00 00 00 xx 	Grid position Width (hex): as horizontal line, left to right.
				pages[i].images[j].gridPosW, _ = readUint32(f)
				// 0034 	4 	00 00 00 xx 	Grid position Height (hex): next position after each horizontal line.
				pages[i].images[j].gridPosH, _ = readUint32(f)
				// 0038 	4 	00 00 00 0x 	Layer level: layer 0 (max zoom) appear first.
				pages[i].images[j].layer, _ = readUint32(f)
				// 003C 	4 	00 00 xx xx 	Length of the image (hex)
				pages[i].images[j].fileLength, _ = readUint32(f)
				// 0040 	4 	00 00 00 xx 	Length padding of the image (hex)
				pages[i].images[j].fileLengthPadding, _ = readUint32(f)
				// 0044 	4 	00 00 00 00 	Not used?
				_, _ = readUint32(f)
				// 0048 	4 	00 00 0x xx 	Width image (hex)
				pages[i].images[j].width, _ = readUint32(f)
				// 004C 	4 	00 00 0x xx 	Height image (hex)
				pages[i].images[j].height, _ = readUint32(f)
			} else {
				panic("not implemented")
			}

			log.Printf("   > %#v", pages[i].images[j])
		}

		// Read BLK
		readCompare(f, []byte{0x42, 0x4C, 0x4B, 0x5F})

		// XXXX 	4 	xx xx xx xx 	Total length embedded images (with FF padding)
		pages[i].lengthImages, _ = readUint32(f)
		log.Printf("[%v] lengthImages: %v", i, pages[i].lengthImages)

		// START IMAGES
		readCompare(f, []byte{00, 00, 00, 02, 00, 00, 00, 00})

		// Skip non images for now.
		if pages[i].imageType != "jpeg" {
			continue
		}

		// Create a new image
		newImage := image.NewRGBA(image.Rect(0, 0, pages[i].imageWidth, pages[i].imageHeight))

		for j := 0; j < numImages; j++ {

			// Skip if not layer 0.
			if pages[i].images[j].layer != 0 {
				_, _ = f.Seek(int64(pages[i].images[j].fileLength+pages[i].images[j].fileLengthPadding), 1)
				continue
			}

			// Load the image.
			rawImage, _ := readBytes(f, pages[i].images[j].fileLength)
			img, err := jpeg.Decode(bytes.NewBuffer(rawImage))
			if err != nil {
				return fmt.Errorf("unable to decode image: %v", j)
			}

			// merge images.
			x := pages[i].images[j].gridPosW * 256
			y := pages[i].images[j].gridPosH * 256
			bounds := img.Bounds()
			draw.Draw(newImage, image.Rect(x, y, x+bounds.Dx(), y+bounds.Dy()), img, bounds.Min, draw.Over)

			// Skip padding.
			_, _ = f.Seek(int64(pages[i].images[j].fileLengthPadding), 1)
		}

		// Save the image
		imgFile, err := os.Create(fmt.Sprintf("out_page_%v.png", i))
		if err != nil {
			return fmt.Errorf("unable to open file: %v", err)
		}
		err = png.Encode(imgFile, newImage)
		if err != nil {
			return fmt.Errorf("unable to encode jpeg: %v", err)
		}
		closeErr := imgFile.Close()
		if closeErr != nil {
			return fmt.Errorf("unable to close output file: %v", err)
		}

	}

	log.Printf(" >> Databases done.")

	return nil
}

func readCompare(f *os.File, b []byte) {
	raw, _ := readBytes(f, len(b))
	if bytes.Compare(raw, b) != 0 {
		log.Panicf("does not compare: %v <> %v", raw, b)
	}
}

func readFileNames(f *os.File) error {
	for i := int(0); i < totalDataEntries; i++ {

		_, err := f.Seek(int64(totalLengthFirstPart+pages[i].offsetFileName), 0)
		if err != nil {
			return fmt.Errorf("unable to seek: %v", err)
		}

		nextName, err := readString(f, int(pages[i].lengthFileName))
		if err != nil {
			return fmt.Errorf("unable to read filename %v : %v", i, err)
		}
		pages[i].fileName = nextName

		log.Printf(" > %v", nextName)
	}

	log.Printf(" >> File names done.")

	return nil
}

func readHeader(f *os.File) error {

	// 0000 8 "TGDT0100"
	expectedHeader := "TGDT0100"

	log.Printf("Checking header %s", expectedHeader)
	TGDHeader, err := readString(f, 8)
	if err != nil {
		return err
	}

	if strings.Compare(TGDHeader, expectedHeader) != 0 {
		return fmt.Errorf("header mismatch")
	} else {
		log.Println("...done")
	}

	// 0008 4 Total data entry (next 0x10) in hex
	totalDataEntries, err = readUint32(f)
	if err != nil {
		return err
	}
	log.Printf("totalDataEntries: %v", totalDataEntries)

	// 000C 4 Total Length first part/start second part (first image id.gvd)
	totalLengthFirstPart, err = readUint32(f)
	if err != nil {
		return err
	}
	log.Printf("totalLengthFirstPart: %v", totalLengthFirstPart)

	pages = make([]PageInfo, totalDataEntries)

	// 0020 xx Repeat for pages
	for i := int(0); i < totalDataEntries; i++ {

		// 0010 4 Offset file name.gvd (without header TGDT0100)
		pages[i].offsetFileName, err = readUint32(f)
		if err != nil {
			return err
		}
		log.Printf("[Page %v] offsetFileName: %v", i, pages[i].offsetFileName)

		// 0014 4 Length file name.gvd (00 is not counted)
		pages[i].lengthFileName, err = readUint32(f)
		if err != nil {
			return err
		}
		log.Printf("[Page %v] lengthFileName: %v", i, pages[i].lengthFileName)

		// 0018 4 Offset Data Base Viewer
		pages[i].offsetDataBaseViewer, err = readUint32(f)
		if err != nil {
			return err
		}
		log.Printf("[Page %v] offsetDataBaseViewer: %v", i, pages[i].offsetDataBaseViewer)

		// 001C 4 Length Data base Viewer file
		pages[i].lengthDataBaseViewer, err = readUint32(f)
		if err != nil {
			return err
		}
		log.Printf("[Page %v] lengthDataBaseViewer: %v", i, pages[i].lengthDataBaseViewer)

	}

	// 0XXX xx Filled with 00 until the first image ID.gvd start

	log.Printf(" >> Header done.")

	return nil
}

func readBytes(f *os.File, len int) ([]byte, error) {
	str := make([]byte, len)
	_, err := f.Read(str)
	if err != nil {
		return []byte(""), err
	}
	return str, nil
}

func readString(f *os.File, len int) (string, error) {
	raw, err := readBytes(f, len)
	return string(raw), err
	// Simulate a null terminated string.
	// return string(raw[:clen(raw)]), err
}

func readUint32(f *os.File) (int, error) {
	raw, err := readBytes(f, 4)
	if err != nil {
		return 0, err
	}
	return int(binary.BigEndian.Uint32(raw)), nil
}

func readUint4(f *os.File) (int, int, error) {
	raw, err := readBytes(f, 1)
	if err != nil {
		return 0, 0, err
	}

	rawByte := int(raw[0])
	upper := rawByte >> 4
	lower := rawByte & 0x0F

	return int(upper), int(lower), nil
}

func clen(n []byte) int {
	for i := 0; i < len(n); i++ {
		if n[i] == 0 {
			return i
		}
	}
	return len(n)
}
