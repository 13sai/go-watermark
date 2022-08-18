package gowatermark

import (
	"errors"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/golang/freetype"
	pkgerr "github.com/pkg/errors"
)

var (
	ErrGifBound     = errors.New("gif: image block is out of bounds")
	ErrExt          = errors.New("ext not be allowed")
	ErrPosition     = errors.New("check x or y ")
	ErrFontNotFound = errors.New("font file not found")
	ErrRGBA         = errors.New("rgba not set")
)

type watermark struct {
	err      error
	file     string
	savePath string
	img      *image.NRGBA
	font     Font
	position
	rgba
}

type Font struct {
	File    string
	Size    int
	Content string
}

type position struct {
	x int
	y int
}

type rgba struct {
	r uint8
	g uint8
	b uint8
	a uint8
}

func New() *watermark {
	return &watermark{}
}

func (w *watermark) From(path string) *watermark {
	w.file = path
	return w
}

func (w *watermark) Font(font Font) *watermark {
	w.font = font
	return w
}

func (w *watermark) Position(x, y int) *watermark {
	w.position = position{x, y}
	return w
}

func (w *watermark) RGBA(r, g, b, a uint8) *watermark {
	w.rgba = rgba{r, g, b, a}
	return w
}

func (w *watermark) To(path string) *watermark {
	if !isExistPath(filepath.Dir(path)) {
		err := createDir(path)
		if err != nil {
			w.err = pkgerr.Wrap(err, "createDir failed")
			return w
		}
	}
	w.savePath = path
	return w
}

func (w *watermark) Do() *watermark {
	if w.rgba.a+w.rgba.b+w.rgba.r+w.rgba.g < 1 {
		w.err = ErrRGBA
		return w
	}
	if w.font.File == "" {
		w.err = ErrFontNotFound
		return w
	}
	file, err := os.Open(w.file)
	if err != nil {
		w.err = pkgerr.Wrap(err, "open file failed")
		return w
	}
	defer file.Close()

	_, ext, err := image.DecodeConfig(file)
	if err != nil {
		w.err = pkgerr.Wrap(err, "decode image file failed")
		return w
	}

	if w.savePath == "" {
		w.savePath = "./sai." + ext
	}

	switch ext {
	case "gif":
		w.gifFontWatermark()
	case "jpg", "jpeg", "png":
		w.fontWatermark(ext)
	default:
		w.err = ErrExt
	}
	return w
}

func (w *watermark) gifFontWatermark() {
	file, err := os.Open(w.file)
	if err != nil {
		w.err = pkgerr.Wrap(err, "open image file failed")
		return
	}
	gifStruct, err := gif.DecodeAll(file)
	if err != nil {
		w.err = pkgerr.Wrap(err, "decode gif file failed")
		return
	}
	gifs := make([]*image.Paletted, 0)
	x := 0
	y := 0
	for k, row := range gifStruct.Image {
		img := image.NewNRGBA(row.Bounds())
		if k == 0 {
			x = img.Bounds().Dx()
			y = img.Bounds().Dy()
		}
		if k == 0 && gifStruct.Image[k+1].Bounds().Dx() > x && gifStruct.Image[k+1].Bounds().Dy() > y {
			w.err = ErrGifBound
			return
		}
		if x == img.Bounds().Dx() && y == img.Bounds().Dy() {
			for y := 0; y < img.Bounds().Dy(); y++ {
				for x := 0; x < img.Bounds().Dx(); x++ {
					img.Set(x, y, row.At(x, y))
				}
			}
			w.img = img
			w.addFont()
			if w.err != nil {
				return
			}
			paletted := image.NewPaletted(row.Bounds(), row.Palette)
			draw.Draw(paletted, row.Bounds(), w.img, image.ZP, draw.Src)
			gifs = append(gifs, paletted)
		} else {
			gifs = append(gifs, row)
		}
	}

	waterImg, err := os.Create(w.savePath)
	if err != nil {
		w.err = pkgerr.Wrap(err, "create file failed")
		return
	}
	defer waterImg.Close()

	gifStruct2 := &gif.GIF{
		Image:     gifs,
		Delay:     gifStruct.Delay,
		LoopCount: gifStruct.LoopCount,
	}
	err = gif.EncodeAll(waterImg, gifStruct2)
	if err != nil {
		w.err = pkgerr.Wrap(err, "gif EncodeAll failed")
		return
	}
}

func (w *watermark) fontWatermark(ext string) {
	file, err := os.Open(w.file)
	if err != nil {
		w.err = pkgerr.Wrap(err, "open file failed")
		return
	}
	var decodeImg image.Image
	if ext == "png" {
		decodeImg, err = png.Decode(file)
	} else {
		decodeImg, err = jpeg.Decode(file)
	}
	if err != nil {
		w.err = pkgerr.Wrap(err, "decode img file failed")
		return
	}
	img := image.NewNRGBA(decodeImg.Bounds())
	for y := 0; y < img.Bounds().Dy(); y++ {
		for x := 0; x < img.Bounds().Dx(); x++ {
			img.Set(x, y, decodeImg.At(x, y))
		}
	}
	w.img = img
	w.addFont()
	if w.err != nil {
		return
	}
	waterImg, err := os.Create(w.savePath)
	if err != nil {
		w.err = pkgerr.Wrap(err, "create save file failed")
		return
	}
	defer waterImg.Close()

	if ext == "png" {
		err = png.Encode(waterImg, img)
	} else {
		err = jpeg.Encode(waterImg, img, &jpeg.Options{100})
	}
	if err != nil {
		w.err = pkgerr.Wrap(err, "encode image file failed")
	}
}

func (w *watermark) addFont() {
	fontBytes, err := ioutil.ReadFile(w.font.File)
	if err != nil {
		w.err = pkgerr.Wrap(err, "ReadFile failed")
		return
	}
	font, err := freetype.ParseFont(fontBytes)
	if err != nil {
		w.err = pkgerr.Wrap(err, "ParseFont failed")
		return
	}

	ctx := freetype.NewContext()
	ctx.SetDPI(56)
	ctx.SetFont(font)
	ctx.SetFontSize(float64(w.font.Size))
	ctx.SetClip(w.img.Bounds())
	ctx.SetDst(w.img)
	ctx.SetSrc(image.NewUniform(color.RGBA{R: w.rgba.r, G: w.rgba.g, B: w.rgba.b, A: w.rgba.a}))

	pt := freetype.Pt(w.position.x, w.position.y)
	_, w.err = ctx.DrawString(w.font.Content, pt)
}

func (w *watermark) Error() error {
	return w.err
}
