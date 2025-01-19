#!/usr/bin/env python3

import sys
import time
import shutil
import cv2

def main():
    cap = cv2.VideoCapture(0)
    if not cap.isOpened():
        print("Could not open webcam.")
        sys.exit(1)

    # Switch to the alternate screen buffer, hide cursor
    print("\033[?1049h\033[?25l", end="")

    try:
        while True:
            ret, frame = cap.read()
            if not ret:
                break

            # Get terminal size (columns = width, rows = height in characters)
            columns, rows = shutil.get_terminal_size()
            
            # We'll reserve 1 row for safety to avoid scrolling
            # and use 'rows - 1' for video lines.
            # Each printed line = 2 lines of image pixels (due to half-block).
            # So the final image height = 2 * (rows - 1).
            term_width = columns
            term_height = max(1, rows - 1) * 2

            # Resize webcam image to fit the terminal dimension
            # Note: The order in cv2.resize() is (width, height).
            frame = cv2.resize(frame, (term_width, term_height), interpolation=cv2.INTER_AREA)

            # Convert BGR to RGB
            frame = cv2.cvtColor(frame, cv2.COLOR_BGR2RGB)

            # Move cursor to top-left without clearing the screen
            print("\033[H", end="")

            # We’ll iterate through the image by pairs of rows:
            # i = 0,2,4,... so top row = i, bottom row = i+1
            # If the resized height is odd, we’ll skip the last row to avoid out of range.
            for i in range(0, term_height - 1, 2):
                top_row = frame[i]
                bottom_row = frame[i + 1]

                # Build a single line of half-block characters
                line_builder = []
                for x in range(term_width):
                    # top pixel (foreground color)
                    rT, gT, bT = top_row[x]
                    # bottom pixel (background color)
                    rB, gB, bB = bottom_row[x]

                    # \033[38;2;R;G;Bm = set foreground color
                    # \033[48;2;R;G;Bm = set background color
                    # '▀' (U+2580) draws the top half of the cell in the foreground color.
                    line_builder.append(
                        f"\033[38;2;{rT};{gT};{bT}m\033[48;2;{rB};{gB};{bB}m▀"
                    )
                
                # Reset at the end of the line
                line_builder.append("\033[0m")
                print("".join(line_builder))

            # Small delay to avoid overloading the CPU/terminal
            time.sleep(0.03)

    except KeyboardInterrupt:
        pass
    finally:
        cap.release()
        # Return to normal screen buffer, show cursor
        print("\033[?1049l\033[?25h", end="")

if __name__ == "__main__":
    main()
