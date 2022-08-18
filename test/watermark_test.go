package test

import (
	"testing"

	"github.com/13sai/gowatermark"
)

func TestWatermark(t *testing.T) {
	wt := gowatermark.New()
	fileName := "go.jpeg"
	FontFile := "/System/Library/Fonts/STHeiti Medium.ttc" //字体路径
	font := gowatermark.Font{FontFile, 16, "hello"}
	err := wt.From(fileName).Font(font).Position(20, 20).RGBA(20, 20, 100, 255).To("gowt.jpeg").Do().Error()
	if err != nil {
		t.Logf("err=%v", err.Error())
	} else {
		t.Log("success")
	}
}
