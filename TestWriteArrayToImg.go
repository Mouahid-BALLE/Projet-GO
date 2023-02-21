package main

import (
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
)

func arrayToImage(pixels [][]int) *image.RGBA {
	height := len(pixels)
	width := len(pixels[0])

	// Create a new image with the correct dimensions
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Iterate through the array and set the pixel values in the image
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			gray := uint8(pixels[y][x])
			img.Set(x, y, color.Gray{gray})
		}
	}

	return img
}

func main() {
	pixels := [][]int{
		{255, 255, 0, 0},
		{0, 0, 255, 0},
		{0, 128, 0, 128},
		{0, 0, 128, 0},
	}

	img := arrayToImage(pixels)

	// Save the image to a file
	file, err := os.Create("image.png")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	err = png.Encode(file, img)
	if err != nil {
		log.Fatal(err)
	}
}

