package test

import (
	"github.com/13sai/gowatermark"
	"testing"
)

func TestWatermark(t *testing.T) {
	wt := gowatermark.New()
	fileName := "go.jpeg"
	FontFile := "/System/Library/Fonts/STHeiti Medium.ttc" //字体路径
	str := gowatermark.FontInfo{24, "sai0556", gowatermark.TopLeft, 20, 20, 100, 100, 88, 255}
	err := wt.From(fileName).Font(FontFile).To("gowt.jpeg").AddWords(str).Do().Error()
	if err != nil {
		t.Logf("err=%v", err.Error())
	} else {
		t.Log("success")
	}
}