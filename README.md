# gowatermark

watermark powered with golang

```sh
go get github.com/13sai/gowatermark
```


```golang
package main

import (
	"fmt"
	"github.com/13sai/gowatermark"
)

func main() {
	wt := gowatermark.New()
	fileName := "go.jpeg"
	FontFile := "/System/Library/Fonts/STHeiti Medium.ttc" //字体路径
	font := gowatermark.Font{FontFile, 16, "sai0556"}
	err := wt.From(fileName).Font(font).Position(20, 20).RGBA(20, 20, 100, 255).To("gowt.jpeg").Do().Error()
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println("success")
	}
}
```