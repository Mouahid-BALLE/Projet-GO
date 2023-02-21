package main

import (
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"log"
	"os"
	"runtime"
	"sync"
	"time"
)

func init() {
	image.RegisterFormat("jpeg", "jpeg", jpeg.Decode, jpeg.DecodeConfig)
	image.RegisterFormat("png", "png", png.Decode, png.DecodeConfig)
}

func imageToArray(img image.Image) [][]float64 {
	bounds := img.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y

	// Initialize the 2D array with the correct dimensions
	pixels := make([][]float64, height)
	for i := range pixels {
		pixels[i] = make([]float64, width)

	}

	// Iterate through the image pixels and populate the array
	for y := bounds.Min.Y; y < height; y++ {
		for x := bounds.Min.X; x < width; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			// convert pixel color to grayscale
			pixels[y][x] = float64(((float64(r) * 0.2126) + (float64(g) * 0.7152) + (float64(b) * 0.0722)) * (255.0 / 65534.0))
		}
	}

	return pixels
}

func arrayToImage(pixels [][]float64) *image.Gray {
	height := len(pixels)
	width := len(pixels[0])

	// Create a new image with the correct dimensions
	img := image.NewGray(image.Rect(0, 0, width, height))

	// Iterate through the array and set the pixel values in the image
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.Gray{uint8(pixels[y][x])})
		}
	}

	return img
}

func threadedDithering(array *[][]float64, nbThread int) {
	var wg sync.WaitGroup
	rowSize := len(*array) / nbThread
	wg.Add(nbThread)
	for i := 0; i < nbThread; i++ {
		start := i * rowSize
		end := (i + 1) * rowSize

		go func(i int) {
			dither(array, start, end)
			wg.Done()
		}(i)
	}
	wg.Wait()
}

func dither(pixels *[][]float64, start, end int) {

	height := len(*pixels)
	width := len((*pixels)[0])

	error_values := make([][]float64, height)
	for i := range *pixels {
		error_values[i] = make([]float64, width)
	}
	var difference = 0.0
	// Iterate through the array and set the pixel values in the image

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			val := (*pixels)[y][x] + error_values[y][x]

			if val > 128.0 {
				difference = val - 128.0
				(*pixels)[y][x] = 255.0 // Write a white pixel
			} else {
				difference = val
				(*pixels)[y][x] = 0.0 // Write a black pixel
			}
			if x+1 < width {
				error_values[y][x+1] = float64((difference * 7.0) / 16.0)
			}
			if y+1 < height && x-1 >= 0 {
				error_values[y+1][x-1] = float64((difference * 3.0) / 16.0)
			}
			if y+1 < height {
				error_values[y+1][x] = float64((difference * 5.0) / 16.0)
			}
			if y+1 < height && x+1 < width {
				error_values[y+1][x+1] = float64((difference * 1.0) / 16.0)

			}

		}
	}

}

func main() {
	log.Print(runtime.NumCPU())
	// Open the original image
	original, err := os.Open("test5.jpg")
	if err != nil {
		log.Fatal(err)
	}
	defer original.Close()
	log.Println("Image Loaded")
	// Decode the image
	img, _, err := image.Decode(original)
	if err != nil {
		log.Fatal(err)
	}

	// Convert the image to a 2D array
	imageArray := imageToArray(img)

	start := time.Now()
	//threadedDithering(&imageArray, 10)
	dither(&imageArray, 0, len(imageArray))
	log.Printf("Dithering took %s", time.Since(start))

	imgOut := arrayToImage(imageArray)

	log.Println("Image Dithered")
	// Save the image to a file
	file, err := os.Create("image2.png")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	err = png.Encode(file, imgOut)
	if err != nil {
		log.Fatal(err)
	}

}
