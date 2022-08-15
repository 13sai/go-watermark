# gowatermark

watermark powered with golang

```sh
go get github.com/13sai/gowatermark
```


```golang
wt := gowatermark.New()
fileName := "go.jpeg"
FontFile := "/System/Library/Fonts/STHeiti Medium.ttc" //字体路径
str := FontInfo{24, "sai0556", TopLeft, 20, 20, 100, 100, 88, 255}
err := wt.From(fileName).Font(FontFile).AddWords(str).Do().Error()
t.Log("err=", err)
```