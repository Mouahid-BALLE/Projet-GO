package main

import (
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"net/http"
	"os"
	"runtime"
	"sync"
	"time"
)

func init() {
	image.RegisterFormat("jpeg", "jpeg", jpeg.Decode, jpeg.DecodeConfig)
	image.RegisterFormat("png", "png", png.Decode, png.DecodeConfig)
}

func main() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/upload", uploadHandler)
	http.HandleFunc("/download", downloadHandler)
	http.ListenAndServe(":8080", nil)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	html := `
	<html>
		<style>
			body {
				background-color: #b8b8b8;
			}
			#Title {
				color: #FFFFFF;
				text-shadow: 4px 3px 0 #7A7A7A;
				color: #FFFFFF;
				background: #b8b8b8;
				font-size: 5vw;
			}
		</style>
		<body>
			<div id="Title">Algorithme de Floyd-steinberg</div> <br><br>
			<form action="/upload" method="post" enctype="multipart/form-data">
				<input type="file" name="image">
				<input type="submit" value="Upload">
			</form>
		</body>
	</html>
	`
	io.WriteString(w, html)
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Print(runtime.NumCPU())
	// Convert the image to a 2D array
	imageArray := imageToArray(img)

	start := time.Now()
	threadedDithering(&imageArray, 10)
	log.Printf("Dithering took %s", time.Since(start))

	imgOut := arrayToImage(imageArray)

	log.Println("Image Dithered")

	tempFile, err := os.Create(header.Filename)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tempFile.Close()

	err = jpeg.Encode(tempFile, imgOut, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	html := `
	<html>
		<style>
			body {
				background-color: #b8b8b8;
		  	}
			#Title {
				color: #FFFFFF;
				text-shadow: 4px 3px 0 #7A7A7A;
				color: #FFFFFF;
				background: #b8b8b8;
				font-size: 5vw;
			}
		</style>
		<body>
			<div id="Title">Algorithme de Floyd-steinberg</div>  <br><br>
			<a href="/download?filename=` + header.Filename + `"> Cliquer pour telecharger l'image! </a>
		</body>
	</html>
	`
	io.WriteString(w, html)
}

func downloadHandler(w http.ResponseWriter, r *http.Request) {
	filename := r.URL.Query().Get("filename")
	if filename == "" {
		http.Error(w, "Invalid filename", http.StatusBadRequest)
		return
	}
	file, err := os.Open(filename)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	w.Header().Set("Content-Disposition", "attachment; filename="+filename)
	w.Header().Set("Content-Type", "image/jpeg")

	io.Copy(w, file)
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
