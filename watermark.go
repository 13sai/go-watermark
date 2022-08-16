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
)

// 水印的位置
const (
	TopLeft int = iota
	TopRight
	BottomLeft
	BottomRight
	Center
)

var (
	ErrGifBound     = errors.New("gif: image block is out of bounds")
	ErrExt          = errors.New("ext not be allowed")
	ErrPosition     = errors.New("check x or y ")
	ErrFontNotFound = errors.New("font file not found")
)

type watermark struct {
	err      error
	file string
	savePath string
	content    FontInfo
	NRGBA    *image.NRGBA
	FontFile string
}

type FontInfo struct {
	Size     float64 // 文字大小
	Content  string  // 文字内容
	Position int     // 文字存放位置
	Dx       int     // 文字x轴留白距离
	Dy       int     // 文字y轴留白距离
	R        uint8
	G        uint8
	B        uint8
	A        uint8
}

func New() *watermark {
	return &watermark{
	}
}

func (w *watermark) From(path string) *watermark {
	w.file = path
	return w
}

func (w *watermark) Font(file string) *watermark {
	w.FontFile = file
	return w
}

func (w *watermark) AddWords(content FontInfo) *watermark {
	w.content = content
	return w
}

func (w *watermark) To(path string) *watermark {
	if !isExistPath(filepath.Dir(path)) {
		err := createDir(path)
		if err != nil {
			w.err = err
			return w
		}
	}
	w.savePath = path
	return w
	
}

func (w *watermark) Do() *watermark {
	if w.FontFile == "" {
		w.err = ErrFontNotFound
		return w
	}
	file, err := os.Open(w.file)
	if err != nil {
		w.err = err
		return w
	}
	defer file.Close()

	_, ext, err := image.DecodeConfig(file)
	if err != nil {
		w.err = err
		return w
	}

	if w.savePath == "" {
		w.savePath = "./sai."+ext
	}
	
	switch ext {
	case "gif":
		w.gifFontWatermark()
	case "jpg", "png", "jpeg":
		w.fontWatermark(ext)
	default:
		w.err = ErrExt
	}
	return w
}

func (w *watermark) gifFontWatermark() {
	file, err := os.Open(w.file)
	if err != nil {
		w.err = err
		return
	}
	gifStruct, err := gif.DecodeAll(file)
	if err != nil {
		w.err = err
		return
	}
	gifs := make([]*image.Paletted, 0)
	x0 := 0
	y0 := 0
	for k, row := range gifStruct.Image {
		NRGBA := image.NewNRGBA(row.Bounds())
		if k == 0 {
			x0 = NRGBA.Bounds().Dx()
			y0 = NRGBA.Bounds().Dy()
		}
		if k == 0 && gifStruct.Image[k+1].Bounds().Dx() > x0 && gifStruct.Image[k+1].Bounds().Dy() > y0 {
			w.err = ErrGifBound
			return
		}
		if x0 == NRGBA.Bounds().Dx() && y0 == NRGBA.Bounds().Dy() {
			for y := 0; y < NRGBA.Bounds().Dy(); y++ {
				for x := 0; x < NRGBA.Bounds().Dx(); x++ {
					NRGBA.Set(x, y, row.At(x, y))
				}
			}
			w.NRGBA = NRGBA
			w.addFont()
			if w.err != nil {
				break
			}
			paletted := image.NewPaletted(row.Bounds(), row.Palette)
			draw.Draw(paletted, row.Bounds(), w.NRGBA, image.ZP, draw.Src)
			gifs = append(gifs, paletted)
		} else {
			gifs = append(gifs, row)
		}
	}
	if err != nil {
		w.err = err
		return
	}
	waterImg, err := os.Create(w.savePath)
	if err != nil {
		w.err = err
		return
	}
	defer waterImg.Close()

	gifStruct2 := &gif.GIF{
		Image:     gifs,
		Delay:     gifStruct.Delay,
		LoopCount: gifStruct.LoopCount,
	}
	w.err = gif.EncodeAll(waterImg, gifStruct2)
	return
}

func (w *watermark) fontWatermark(ext string) {
	file, err := os.Open(w.file)
	if err != nil {
		w.err = err
		return
	}
	var staticImg image.Image
	if ext == "png" {
		staticImg, w.err = png.Decode(file)
	} else {
		staticImg, w.err = jpeg.Decode(file)
	}
	if w.err != nil {
		return
	}
	img := image.NewNRGBA(staticImg.Bounds())
	for y := 0; y < img.Bounds().Dy(); y++ {
		for x := 0; x < img.Bounds().Dx(); x++ {
			img.Set(x, y, staticImg.At(x, y))
		}
	}
	w.NRGBA = img
	w.addFont()
	if w.err != nil {
		return
	}
	waterImg, err := os.Create(w.savePath)
	if err != nil {
		w.err = err
		return
	}
	defer waterImg.Close()

	if ext == "png" {
		w.err = png.Encode(waterImg, img)
	} else {
		w.err = jpeg.Encode(waterImg, img, &jpeg.Options{100})
	}
	if w.err != nil {
		return
	}
	return
}

func (w *watermark) addFont() {
	var err error
	fontBytes, err := ioutil.ReadFile(w.FontFile)
	if err != nil {
		w.err = err
		return
	}
	font, err := freetype.ParseFont(fontBytes)
	if err != nil {
		w.err = err
		return
	}

	info := w.content.Content
	f := freetype.NewContext()
	f.SetDPI(108)
	f.SetFont(font)
	f.SetFontSize(w.content.Size)
	f.SetClip(w.NRGBA.Bounds())
	f.SetDst(w.NRGBA)

	f.SetSrc(image.NewUniform(color.RGBA{R: w.content.R, G: w.content.G, B: w.content.B, A: w.content.A}))
	x, y := 0, 0
	switch int(w.content.Position) {
	case TopLeft:
		x = w.content.Dx
		y = w.content.Dy + int(f.PointToFixed(w.content.Size)>>6)
	case TopRight:
		x = w.NRGBA.Bounds().Dx() - len(info)*10 - w.content.Dx
		y = w.content.Dy + int(f.PointToFixed(w.content.Size)>>6)
	case BottomLeft:
		x = w.content.Dx
		y = w.NRGBA.Bounds().Dy() - w.content.Dy
	case BottomRight:
		x = w.NRGBA.Bounds().Dx() - len(info)*10 - w.content.Dx
		y = w.NRGBA.Bounds().Dy() - w.content.Dy
	case Center:
		x = (w.NRGBA.Bounds().Dx() - len(info)*10) / 2
		y = (w.NRGBA.Bounds().Dy() - w.content.Dy) / 2
	default:
		w.err = ErrPosition
		return
	}
	pt := freetype.Pt(x, y)
	_, w.err = f.DrawString(info, pt)
}

func (w *watermark) Error() error {
	return w.err
}
