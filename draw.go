package main

import (
	/*
		 #cgo LDFLAGS: -L./ -L/usr/local/freetype/lib -lfreetype -lDrawBrush -lstdc++
		 #cgo CFLAGS: -I./ -I/usr/local/freetype/include/freetype2
		 #include "DrawBrush.h"
		typedef struct FTLibrary ft_library;
		typedef struct RGBA Color32;

		static void CreateRGBA(struct RGBA *r,uint8 ri, uint8 gi, uint8 bi, uint8 ai)
		{
			InitRGBA(r,ri,gi,bi,ai);

		}

		static void Free(struct Result *r)
		{
			free(r->data);

		}

		typedef struct Result Result;

	*/
	"C"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"reflect"
	"unsafe"
	"bufio"
	// "flag"
	"image/png"
	"log"
	"os"
	"strings"
	"strconv"
	"golang.org/x/image/math/fixed"
)


type DrawFont struct {
	library   C.ft_library
	sideScale int //被除数
}

func C2GOUint8(buf *C.uint8, size int) (ret []uint8) {
	hdr := (*reflect.SliceHeader)((unsafe.Pointer(&ret)))
	hdr.Cap = size
	hdr.Len = size
	hdr.Data = uintptr(unsafe.Pointer(buf))
	return
}

func NewDrawFont(fontName string) *DrawFont {
	d := &DrawFont{sideScale: 16}
	font := []byte(fontName)
	r := C.InitFTLibrary((*C.char)(unsafe.Pointer(&font[0])), &d.library)
	if int(r) == 1 {
		fmt.Println("InitFTLibrary OK")
		return d
	} else {
		fmt.Println("InitFTLibrary faild")
		return nil
	}
}


func (d *DrawFont) FreeDraw() {
	C.FreeLibrary(&d.library)
}

func (d *DrawFont) PointToFixed(x float64) fixed.Int26_6 {
	return fixed.Int26_6(x * 64.0)
}

func (d *DrawFont) DrawText(text rune, fontColor color.RGBA, fontsize int) *image.RGBA {

	var scolor, fcolor C.Color32

	C.CreateRGBA(&scolor, 0, 0, 0, 0)
	C.CreateRGBA(&fcolor, C.uint8(fontColor.R), C.uint8(fontColor.G), C.uint8(fontColor.B), C.uint8(fontColor.A))

	var data C.Result
	C.WriteGlyph(&d.library, C.ulong(text), C.int(fontsize), &fcolor, &scolor, 0.0, &data)

	return &image.RGBA{
		Pix:    C2GOUint8(data.data, int(data.len)),
		Stride: int(data.width * 4),
		Rect:   image.Rect(0, 0, int(data.width), int(data.height)),
	}

}

func (d *DrawFont) DrawTextWithOutLine(text rune, fontColor, outLineColor color.RGBA, fontsize, outlineSize float64) *image.RGBA {

	var scolor, fcolor C.Color32

	C.CreateRGBA(&scolor, C.uint8(outLineColor.R), C.uint8(outLineColor.G), C.uint8(outLineColor.B), C.uint8(outLineColor.A))
	C.CreateRGBA(&fcolor, C.uint8(fontColor.R), C.uint8(fontColor.G), C.uint8(fontColor.B), C.uint8(fontColor.A))

	var data C.Result
	C.WriteGlyph(&d.library,
		C.ulong(text), C.int(fontsize), &fcolor, &scolor, C.float(outlineSize), &data,
	)

	return &image.RGBA{
		Pix:    C2GOUint8(data.data, int(data.len)),
		Stride: int(data.width * 4),
		Rect:   image.Rect(0, 0, int(data.width), int(data.height)),
	}
}

func (d *DrawFont) DrawStringWithOutLine(text string, fontColor, outLineColor color.RGBA, fontsize, outlineSize float64) *image.RGBA {

	var scolor, fcolor C.Color32

	C.CreateRGBA(&scolor, C.uint8(outLineColor.R), C.uint8(outLineColor.G), C.uint8(outLineColor.B), C.uint8(outLineColor.A))
	C.CreateRGBA(&fcolor, C.uint8(fontColor.R), C.uint8(fontColor.G), C.uint8(fontColor.B), C.uint8(fontColor.A))

	var img *image.RGBA

	var list []*image.RGBA
	var offsetx int = 0
	var height int = 0
	var width int = 0
	var dlist []C.Result
	for _, s := range []rune(text) {
		var data C.Result
		C.WriteGlyph(&d.library,
			C.ulong(s), C.int(fontsize), &fcolor, &scolor, C.float(outlineSize), &data,
		)
		f := &image.RGBA{
			Pix:    C2GOUint8(data.data, int(data.len)),
			Stride: int(data.width * 4),
			Rect:   image.Rect(0, 0, int(data.width), int(data.height)),
		}
		list = append(list, f)

		dlist = append(dlist, data)

		if width < int(data.width) {
			width = int(data.width)
		}
		if height < int(data.height) {
			height = int(data.height)
		}
	}

	// fmt.Println("width:", int(width))

	if img == nil {
		img = image.NewRGBA(image.Rect(0, 0, (int(width)+int(fontsize)/d.sideScale)*len([]rune(text)), int(height)))
	}

	for _, f := range list {

		sr := f.Bounds()                                                       // 获取要复制图片的尺寸 (height - f.Bounds().Dy()) / 2)
		r := sr.Sub(sr.Min).Add(image.Pt(offsetx, (height-f.Bounds().Dy())/2)) // 目标图的要剪切区域

		offsetx += width + int(fontsize)/d.sideScale
		draw.Draw(img, r, f, sr.Min, draw.Src)
	}

	for _, f := range dlist {

		C.Free(&f)
	}
	return img
}

//export DrawStringToImg
func DrawStringToImg(text_c,fontfile_c,imgfile_c,savefile_c *C.char,fontsize,outsize float64,fontcolor_c,outcolor_c *C.char,pt_w,pt_h int) {
	// text,fontfile,imgfile,savefile,fontcolor,outcolor

	text := C.GoString(text_c)
	fontfile := C.GoString(fontfile_c)
	imgfile := C.GoString(imgfile_c)
	savefile := C.GoString(savefile_c)
	fontcolor := C.GoString(fontcolor_c)
	outcolor := C.GoString(outcolor_c)

	b1, err := os.Open(imgfile)

	if err != nil {
		panic(err)
	}

	rgba, _, err := image.Decode(b1)
	if err != nil {
		panic(err)
	}

	font := NewDrawFont(fontfile)//"./wryhBold.ttf"

	fc := strings.Split(fontcolor, ",")
	oc := strings.Split(outcolor, ",")
	fc_0, _ := strconv.ParseInt(fc[0], 10, 64)
	fc_1, _ := strconv.ParseInt(fc[1], 10, 64)
	fc_2, _ := strconv.ParseInt(fc[2], 10, 64)
	fc_3, _ := strconv.ParseInt(fc[3], 10, 64)
	oc_0, _ := strconv.ParseInt(oc[0], 10, 64)
	oc_1, _ := strconv.ParseInt(oc[1], 10, 64)
	oc_2, _ := strconv.ParseInt(oc[2], 10, 64)
	oc_3, _ := strconv.ParseInt(oc[3], 10, 64)

	FG := color.RGBA{uint8(fc_0),uint8(fc_1),uint8(fc_2),uint8(fc_3)}
	// yellowG := color.RGBA{0xff, 241, 0, 0xff}
	outG := color.RGBA{uint8(oc_0),uint8(oc_1),uint8(oc_2),uint8(oc_3)}

	

	var f *image.RGBA
	
	f = font.DrawStringWithOutLine(text, FG, outG, fontsize, outsize)

	pt := image.Pt(pt_w,pt_h)
	// int((rgba.Bounds().Dx()-f.Bounds().Dx())/2),
	// int((rgba.Bounds().Dy()-(f.Bounds().Dy()/5+f.Bounds().Dy()))/2)
	sr := f.Bounds()            // 获取要复制图片的尺寸
	r := sr.Sub(sr.Min).Add(pt) // 目标图的要剪切区域

	draw.Draw(rgba.(draw.Image), r, f, sr.Min, draw.Over)
	
	font.FreeDraw()

	// Save that RGBA image to disk.
	outFile, err := os.Create(savefile)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	defer outFile.Close()
	b := bufio.NewWriter(outFile)
	err = png.Encode(b, rgba)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	err = b.Flush()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	fmt.Println("Wrote out.png OK.")
}

func main(){
	// DrawStringToImg("测试","wryhBold.ttf","background.png","out.png",100,20,"220,20,60,0","27,91,97,0",200,200)
}