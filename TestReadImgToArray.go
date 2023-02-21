package main


import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"log"
	"os"
)

func init() {
	image.RegisterFormat("jpeg", "jpeg", jpeg.Decode, jpeg.DecodeConfig)
	image.RegisterFormat("png", "png", png.Decode, png.DecodeConfig)
}

func imageToArray(img image.Image) [][]int {
	bounds := img.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y

	// Initialize the 2D array with the correct dimensions
	pixels := make([][]int, height)
	for i := range pixels {
		pixels[i] = make([]int, width)

	}
	var m, n, p = bounds.Max.Y, bounds.Max.X, 4
	buf := make([]int, m*n*p)
	pix := make([][][]int, m)
	for i := range pix {
		pix[i] = make([][]int, n)
		for j := range pix[i] {
			pix[i][j] = buf[:p:p]
			buf = buf[p:]
		}
	}

	// Iterate through the image pixels and populate the array
	for y := bounds.Min.Y; y < height; y++ {
		for x := bounds.Min.X; x < width; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			pix[y][x][0] = int(r)
			pix[y][x][1] = int(g)
			pix[y][x][2] = int(b)
			// convert pixel color to grayscale
			pixels[y][x] = int((float64(r) * 0.299) + (float64(g) * 0.587) + (float64(b) * 0.114))
		}
	}

	return pixels
}

func main() {
	// Open the original image
	original, err := os.Open("image.png")
	if err != nil {
		log.Fatal(err)
	}
	defer original.Close()

	// Decode the image
	img, _, err := image.Decode(original)
	if err != nil {
		log.Fatal(err)
	}

	// Convert the image to a 2D array
	imageArray := imageToArray(img)
	fmt.Print(imageArray)

}
