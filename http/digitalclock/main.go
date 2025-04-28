//go:build !solution

package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	symbolMap     map[rune]string
	digitWidth    int
	colonWidth    int
	symbolHeight  int
	preCalculated bool
)

func calculateSymbolDimensions() {
	if preCalculated {
		return
	}
	symbolHeight = len(strings.Split(Zero, "\n"))

	if symbolHeight > 0 {
		firstLine := strings.Split(Zero, "\n")[0]
		digitWidth = len(firstLine)
	}

	if symbolHeight > 0 {
		firstLineColon := strings.Split(Colon, "\n")[0]
		colonWidth = len(firstLineColon)
	}

	symbolMap = map[rune]string{
		'0': Zero,
		'1': One,
		'2': Two,
		'3': Three,
		'4': Four,
		'5': Five,
		'6': Six,
		'7': Seven,
		'8': Eight,
		'9': Nine,
		':': Colon,
	}
	preCalculated = true

	if symbolHeight == 0 || digitWidth == 0 || colonWidth == 0 {
		log.Fatal("Could not calculate symbol dimensions from symbols.go")
	}
}

func sendError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(statusCode)
	fmt.Fprintln(w, message)
}

func generateClockImage(timeStr string, k int) (image.Image, error) {
	calculateSymbolDimensions()

	totalWidth := (6*digitWidth + 2*colonWidth) * k
	totalHeight := symbolHeight * k

	img := image.NewRGBA(image.Rect(0, 0, totalWidth, totalHeight))

	currentXOffset := 0

	for _, char := range timeStr {
		symbolStr, ok := symbolMap[char]
		if !ok {
			return nil, fmt.Errorf("invalid character in time string: %c", char)
		}

		symbolLines := strings.Split(symbolStr, "\n")
		currentSymbolWidth := digitWidth
		if char == ':' {
			currentSymbolWidth = colonWidth
		}

		if len(symbolLines) != symbolHeight {
			return nil, fmt.Errorf("symbol height mismatch for %c", char)
		}

		for sy := 0; sy < symbolHeight; sy++ {
			line := symbolLines[sy]
			if len(line) != currentSymbolWidth {
				return nil, fmt.Errorf("symbol width mismatch for %c in line %d", char, sy)
			}
			for sx := 0; sx < currentSymbolWidth; sx++ {
				pixelChar := line[sx]
				var clr color.Color = color.White
				if pixelChar == '1' {
					clr = Cyan
				} else if pixelChar != '.' {
					return nil, fmt.Errorf("invalid character '%c' in symbol definition for %c", pixelChar, char)
				}

				startX := (currentXOffset + sx) * k
				startY := sy * k
				for ky := 0; ky < k; ky++ {
					for kx := 0; kx < k; kx++ {
						img.Set(startX+kx, startY+ky, clr)
					}
				}
			}
		}
		currentXOffset += currentSymbolWidth
	}

	return img, nil
}

func handleClockRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendError(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	params := r.URL.Query()

	kStr := params.Get("k")
	k := 1
	var err error
	if kStr != "" {
		k, err = strconv.Atoi(kStr)
		if err != nil {
			sendError(w, "invalid k: not an integer", http.StatusBadRequest)
			return
		}
		if k < 1 || k > 30 {
			sendError(w, "invalid k: must be between 1 and 30", http.StatusBadRequest)
			return
		}
	}

	timeStr := params.Get("time")
	if timeStr == "" {
		timeStr = time.Now().Format("15:04:05")
	} else {
		_, err := time.Parse("15:04:05", timeStr)
		if err != nil {
			sendError(w, "invalid time format: expected hh:mm:ss", http.StatusBadRequest)
			return
		}
		if len(timeStr) != 8 || timeStr[2] != ':' || timeStr[5] != ':' {
			sendError(w, "invalid time format: expected hh:mm:ss", http.StatusBadRequest)
			return
		}
	}

	img, err := generateClockImage(timeStr, k)
	if err != nil {
		log.Printf("Error generating image: %v", err)
		sendError(w, "internal server error generating image", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.WriteHeader(http.StatusOK)
	err = png.Encode(w, img)
	if err != nil {
		log.Printf("Error encoding or writing png response: %v", err)
	}
}

func main() {
	port := flag.Int("port", 8080, "Port number for the HTTP server")
	flag.Parse()

	calculateSymbolDimensions()

	mux := http.NewServeMux()
	mux.HandleFunc("/", handleClockRequest)

	serverAddr := fmt.Sprintf(":%d", *port)
	log.Printf("Starting digital clock server on %s", serverAddr)

	err := http.ListenAndServe(serverAddr, mux)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
		os.Exit(1)
	}
}
