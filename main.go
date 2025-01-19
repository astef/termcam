package main

import (
	"fmt"
	"image"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"gocv.io/x/gocv"
	"golang.org/x/term"
)

func main() {
	var v any
	var arg string
	if len(os.Args) > 1 {
		arg = os.Args[1]
	}
	if argi, err := strconv.Atoi(arg); err != nil {
		v = arg
	} else {
		v = argi
	}

	// open webcam
	webcam, err := gocv.OpenVideoCapture(v)
	if err != nil {
		log.Fatalf("Error opening webcam: %v", err)
	}
	defer webcam.Close()

	// prepare a Mat to receive frames.
	frame := gocv.NewMat()
	defer frame.Close()

	// switch to alternate screen buffer and hide the cursor.
	// this ensures that when we exit, we go back to the normal terminal.
	fmt.Print("\033[?1049h\033[?25l")

	// handle Ctrl+C or other signals to restore the terminal on exit.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		restoreTerminal()
		os.Exit(0)
	}()

	for {
		// read a frame from the webcam.
		if ok := webcam.Read(&frame); !ok || frame.Empty() {
			fmt.Println("error reading from webcam: empty frame")
			continue
		}

		// get current terminal size (width = columns, height = rows).
		width, height, err := term.GetSize(int(os.Stdin.Fd()))
		if err != nil {
			// fallback to some default if we can't get size.
			width, height = 80, 24
		}

		// we'll leave 1 row for safety to avoid scrolling.
		// each printed line uses two image rows (top + bottom).
		// so the final image height = 2 * (height - 1).
		termWidth := width
		termHeight := (height - 1) * 2
		if termHeight < 2 {
			termHeight = 2
		}

		// resize the frame to (termWidth x termHeight).
		// gocv.Resize expects (dst, src, size, fx, fy, interpolation).
		resized := gocv.NewMat()
		gocv.Resize(frame, &resized, image.Pt(termWidth, termHeight), 0, 0, gocv.InterpolationArea)
		// convert BGR to RGB so we can print correct color.
		gocv.CvtColor(resized, &resized, gocv.ColorBGRToRGB)

		// move cursor to top-left without clearing screen (reduce flicker).
		fmt.Print("\033[H")

		// render using half-block. Each row of output = 2 rows of pixels.
		// we'll iterate i by steps of 2.
		for i := 0; i < termHeight; i += 2 {
			topRow := resized.RowRange(i, i+1)
			botRow := resized.RowRange(i+1, i+2)

			// we'll build a single string for that line, then print.
			line := make([]byte, 0, termWidth*50) // rough capacity guess

			// each pixel has 3 channels (R, G, B).
			for x := 0; x < termWidth; x++ {
				// index for top pixel:
				rT := topRow.GetUCharAt(0, x*3+0)
				gT := topRow.GetUCharAt(0, x*3+1)
				bT := topRow.GetUCharAt(0, x*3+2)

				// index for bottom pixel:
				rB := botRow.GetUCharAt(0, x*3+0)
				gB := botRow.GetUCharAt(0, x*3+1)
				bB := botRow.GetUCharAt(0, x*3+2)

				// \033[38;2;R;G;Bm = set Fg color
				// \033[48;2;R;G;Bm = set Bg color
				// '▀' (Unicode 0x2580) draws top half; background color is below it.
				line = append(line,
					fmt.Sprintf("\033[38;2;%d;%d;%dm\033[48;2;%d;%d;%dm▀",
						rT, gT, bT, rB, gB, bB)...)
			}
			// reset color at line end.
			line = append(line, []byte("\033[0m")...)
			fmt.Println(string(line))

			topRow.Close()
			botRow.Close()
		}

		resized.Close()

		// delay to avoid overwhelming the CPU/terminal.
		time.Sleep(30 * time.Millisecond)
	}
}

func restoreTerminal() {
	// switch back from alternate screen buffer, show cursor again.
	fmt.Print("\033[?1049l\033[?25h")
}
