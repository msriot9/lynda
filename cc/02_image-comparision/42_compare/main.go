package main

import (
	"fmt"
	stdimage "image"
	"image/jpeg"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"math"
)

// at terminal:
// go run -race main.go

type pixel struct {
	r, g, b, a uint32
}

type image struct {
	name   string
	pixels []pixel
	width  int
	height int
}

func main() {
	start := time.Now()

	images, err := getImages()
	if err != nil {
		log.Println("Error getting images", err)
	}

	// compare the images
	matches := compare(images)

	for k, v := range matches {
		fmt.Println("This picture", k, "was in this picture", v)
	}

	fmt.Printf("%.2fs elapsed\n", time.Since(start).Seconds())
}

func getImages() ([]image, error) {

	paths, err := getPaths()
	if err != nil {
		log.Println("Error getting paths", err)
	}

	var mu sync.Mutex
	var wg sync.WaitGroup
	wg.Add(len(paths))

	var images []image
	for _, path := range paths {
		go func(path string) {
			image := getPixels(path)

			mu.Lock()
			{
				images = append(images, image)
			}
			mu.Unlock()

			wg.Done()
		}(path)
	}

	wg.Wait()

	return images, nil
}

func getPaths() ([]string, error) {
	const dir = "../00_images/"
	var paths []string

	wf := func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		paths = append(paths, path)
		return nil
	}

	if err := filepath.Walk(dir, wf); err != nil {
		return nil, err
	}

	return paths, nil
}

func getPixels(path string) image {
	img := loadImage(path)
	bounds := img.Bounds()
	fmt.Println(bounds.Dx(), " x ", bounds.Dy()) // debugging
	pixels := make([]pixel, bounds.Dx()*bounds.Dy())

	for i := 0; i < bounds.Dx()*bounds.Dy(); i++ {
		x := i % bounds.Dx()
		y := i / bounds.Dx()
		r, g, b, a := img.At(x, y).RGBA()
		pixels[i].r = r
		pixels[i].g = g
		pixels[i].b = b
		pixels[i].a = a
	}

	xs := strings.Split(path, "/")
	name := xs[(len(xs) - 1):][0]
	image := image{
		name:   name,
		pixels: pixels,
		width:  bounds.Dx(),
		height: bounds.Dy(),
	}
	return image
}

func loadImage(filename string) stdimage.Image {
	f, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	img, err := jpeg.Decode(f)
	if err != nil {
		log.Fatal(err)
	}

	return img
}

func compare(images []image) map[string]string {

	matches := make(map[string]string)

	// for each image, compare it to the other images
	for _, needle := range images {
		for _, haystack := range images {
			// needle is greater than haystack in width
			// therefore, needle isn't from haystack
			if needle.width > haystack.width {
				continue
			}
			// don't compare with self
			if needle.name == haystack.name {
				continue
			}

			// compareColor
			for k, v := range needle.pixels {
				var diff int64
				diff += int64(math.Abs(float64(v.r - haystack.pixels[k].r)))
				diff += int64(math.Abs(float64(v.g - haystack.pixels[k].g)))
				diff += int64(math.Abs(float64(v.b - haystack.pixels[k].b)))
				diff += int64(math.Abs(float64(v.a - haystack.pixels[k].a)))
				fmt.Printf("%d\n", diff)
			}
//			fmt.Printf("%d\n", int(diff)/(needle.width*needle.height))

			// DEBUGGING - line below is for testing - needs to be changed
			matches[needle.name] = haystack.name
		}
	}
	return matches
}