package render

import (
	"image"
	"image/draw"
)

// FromImage receives an image.Image and a width and returns a pointer to a frame that contains the image in ascii
func FromImage(img image.Image, width int) *Frame {
	img_rect := img.Bounds().Bounds()
	img_h := img_rect.Dy()
	img_w := img_rect.Dx()

	nrgba, ok := img.(*image.NRGBA)
	if !ok {
		nrgba = image.NewNRGBA(img_rect)
		draw.Draw(nrgba, img_rect, img, img_rect.Min, draw.Src)
	}

	// how many real pixels correspond to one character
	block_w := float64(img_w) / float64(width)
	block_h := block_w * 2.4 // ratio of width to height 1:2.4
	height := int(float64(img_h) / block_h)

	ret_frame := NewFrame(width, height)

	for i := range height {
		for j := range width {

			// pixel ranges of THIS character
			start_x := int(float64(j) * block_w)
			end_x := int(float64(j+1) * block_w)

			start_y := int(float64(i) * block_h)
			end_y := int(float64(i+1) * block_h)

			if end_x <= start_x {
				end_x = start_x + 1
			}
			if end_y <= start_y {
				end_y = start_y + 1
			}

			// fix out-of-bounds
			if end_x > img_w {
				end_x = img_w
			}
			if end_y > img_h {
				end_y = img_h
			}

			// accumulate average
			var sum float32
			var count int

			for y := start_y; y < end_y; y++ {
				for x := start_x; x < end_x; x++ {
					// PixOffset takes absolute coordinates, hence the + Min
					off := nrgba.PixOffset(x+img_rect.Min.X, y+img_rect.Min.Y)
					r, g, b := nrgba.Pix[off], nrgba.Pix[off+1], nrgba.Pix[off+2]
					sum += float32(int(r)+int(g)+int(b)) / 3
					count++
				}
			}

			avg := sum / float32(count)
			avg /= 255 // Pix bytes go from 0 to 255

			ret_frame.Set(j, i, avg)
		}
	}

	return ret_frame
}
