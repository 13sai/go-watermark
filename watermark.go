package gowatermark

import (
	"errors"
	"flag"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"os"

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

// 字体路径
var (
	fontFile = flag.String("font", "", "fontfile")
	savePath = flag.String("save", "./", "savepath for watermark")
)

var (
	ErrGifBound     = errors.New("gif: image block is out of bounds")
	ErrExt          = errors.New("ext not be allowed")
	ErrPosition     = errors.New("check x or y ")
	ErrFontNotFound = errors.New("font file not found")
)

type watermark struct {
	Err      error
	FileName string
	SavePath string
	Words    FontInfo
	FileOS   *os.File
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
		SavePath: *savePath,
	}
}

func (w *watermark) From(path string) *watermark {
	w.FileName = path
	return w
}

func (w *watermark) Font(file string) *watermark {
	w.FontFile = file
	return w
}

func (w *watermark) AddWords(words FontInfo) *watermark {
	w.Words = words
	return w
}

func (w *watermark) To(path string) *watermark {
	dir, err := createDir(path)
	if err != nil {
		w.Err = err
		return w
	}
	w.SavePath = dir
	return w
}

func (w *watermark) Do() *watermark {
	if w.FontFile == "" {
		w.Err = ErrFontNotFound
		return w
	}
	file, err := os.Open(w.FileName)
	if err != nil {
		w.Err = err
		return w
	}
	defer w.FileOS.Close()

	_, ext, err := image.DecodeConfig(file)
	if err != nil {
		w.Err = err
		return w
	}

	w.FileOS, w.Err = os.Open(w.FileName)
	switch ext {
	case "gif":
		w.gifFontWatermark()
	case "jpg", "png", "jpeg":
		w.fontWatermark(ext)
	default:
		w.Err = ErrExt
	}
	return w
}

func (w *watermark) gifFontWatermark() {
	gifStruct, err := gif.DecodeAll(w.FileOS)
	if err != nil {
		w.Err = err
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
			w.Err = ErrGifBound
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
			if w.Err != nil {
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
		w.Err = err
		return
	}
	waterImg, err := os.Create(w.SavePath + "wt_" + w.FileOS.Name())
	if err != nil {
		w.Err = err
		return
	}
	defer waterImg.Close()

	gifStruct2 := &gif.GIF{
		Image:     gifs,
		Delay:     gifStruct.Delay,
		LoopCount: gifStruct.LoopCount,
	}
	w.Err = gif.EncodeAll(waterImg, gifStruct2)
	return
}

func (w *watermark) fontWatermark(ext string) {
	var staticImg image.Image
	if ext == "png" {
		staticImg, w.Err = png.Decode(w.FileOS)
	} else {
		staticImg, w.Err = jpeg.Decode(w.FileOS)
	}
	if w.Err != nil {
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
	if w.Err != nil {
		return
	}
	waterImg, err := os.Create(w.SavePath + "wt_" + w.FileOS.Name())
	if err != nil {
		w.Err = err
		return
	}
	defer waterImg.Close()

	if ext == "png" {
		w.Err = png.Encode(waterImg, img)
	} else {
		w.Err = jpeg.Encode(waterImg, img, &jpeg.Options{100})
	}
	if w.Err != nil {
		return
	}
	return
}

func (w *watermark) addFont() {
	var err error
	fontBytes, err := ioutil.ReadFile(w.FontFile)
	if err != nil {
		w.Err = err
		return
	}
	font, err := freetype.ParseFont(fontBytes)
	if err != nil {
		w.Err = err
		return
	}

	info := w.Words.Content
	f := freetype.NewContext()
	f.SetDPI(108)
	f.SetFont(font)
	f.SetFontSize(w.Words.Size)
	f.SetClip(w.NRGBA.Bounds())
	f.SetDst(w.NRGBA)

	f.SetSrc(image.NewUniform(color.RGBA{R: w.Words.R, G: w.Words.G, B: w.Words.B, A: w.Words.A}))
	x, y := 0, 0
	switch int(w.Words.Position) {
	case TopLeft:
		x = w.Words.Dx
		y = w.Words.Dy + int(f.PointToFixed(w.Words.Size)>>6)
	case TopRight:
		x = w.NRGBA.Bounds().Dx() - len(info)*4 - w.Words.Dx
		y = w.Words.Dy + int(f.PointToFixed(w.Words.Size)>>6)
	case BottomLeft:
		x = w.Words.Dx
		y = w.NRGBA.Bounds().Dy() - w.Words.Dy
	case BottomRight:
		x = w.NRGBA.Bounds().Dx() - len(info)*4 - w.Words.Dx
		y = w.NRGBA.Bounds().Dy() - w.Words.Dy
	case Center:
		x = (w.NRGBA.Bounds().Dx() - len(info)*4) / 2
		y = (w.NRGBA.Bounds().Dy() - w.Words.Dy) / 2
	default:
		w.Err = ErrPosition
		return
	}
	pt := freetype.Pt(x, y)
	_, w.Err = f.DrawString(info, pt)
}

func (w *watermark) Error() error {
	return w.Err
}
