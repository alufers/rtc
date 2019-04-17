package main

// #cgo LDFLAGS: -lXinerama -lX11 -lXext
// #cgo pkg-config: cairo
// #include <assert.h>
// #include <stdio.h>
// #include <cairo.h>
// #include <cairo-xlib.h>
// #include <X11/Xlib.h>
// #include <X11/extensions/Xinerama.h>
// #include <X11/extensions/XShm.h>
// #include <X11/Xutil.h>
//Drawable DefRootWinFunc(Display *display) {
// return DefaultRootWindow(display);
// }
//unsigned long XGetPixelFunc( XImage *ximage, int x, int y) {
// return XGetPixel(ximage, x, y);
// }
//int XDestroyImageFunc( XImage *ximage) {
// return XDestroyImage(ximage);
// }
// void ReplaceBG(XImage *ximage) {
// 	for(int i = 0; i < ximage->width * ximage->height * 4; i+=4) {
// 		char tmp = ximage->data[i];
// 		ximage->data[i] = ximage->data[i+2];
// 		ximage->data[i+2] = tmp;
// 	}
// }
import "C"
import (
	"errors"
	"image"
	"unsafe"
)

type ScreenInfo struct {
	ScreenNumber int
	XOrg         int
	YOrg         int
	Width        int
	Height       int
}

func getMonitors() ([]*ScreenInfo, error) {
	disp := C.XOpenDisplay(nil)
	if disp == nil {
		return nil, errors.New("failed to open XDisplay")
	}
	defer C.XCloseDisplay(disp)
	var nmonitors C.int
	monitors := C.XineramaQueryScreens(disp, &nmonitors)
	if monitors == nil {
		return nil, errors.New("failed to query Xinerama screens")
	}
	defer C.XFree(unsafe.Pointer(monitors))
	info := []*ScreenInfo{}
	for i := 0; i < int(nmonitors); i++ {
		scrInfo := (*C.XineramaScreenInfo)(unsafe.Pointer(uintptr(unsafe.Pointer(monitors)) + unsafe.Sizeof(C.XineramaScreenInfo{})*uintptr(i)))
		info = append(info, &ScreenInfo{
			ScreenNumber: int(scrInfo.screen_number),

			XOrg: int(scrInfo.x_org),
			YOrg: int(scrInfo.y_org),

			Width:  int(scrInfo.width),
			Height: int(scrInfo.height),
		})
	}
	return info, nil
}

type LinuxScreenCapturer struct {
	info *ScreenInfo
	disp *C.Display
	xscr C.int
}

func NewCapturer(info *ScreenInfo) (*LinuxScreenCapturer, error) {
	lsc := &LinuxScreenCapturer{
		info: info,
	}
	lsc.disp = C.XOpenDisplay(nil)
	if lsc.disp == nil {
		return nil, errors.New("failed to open XDisplay")
	}
	lsc.xscr = C.XDefaultScreen(lsc.disp)
	return lsc, nil
}

func (lsc *LinuxScreenCapturer) Capture() (image.Image, error) {
	ximage := C.XGetImage(lsc.disp, C.DefRootWinFunc(lsc.disp), C.int(lsc.info.XOrg), C.int(lsc.info.YOrg), C.uint(lsc.info.Width), C.uint(lsc.info.Height), C.AllPlanes, C.ZPixmap)
	if ximage == nil {
		return nil, errors.New("failed to grab image")
	}
	defer C.XDestroyImageFunc(ximage)
	width := int(ximage.width)
	height := int(ximage.height)
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	C.ReplaceBG(ximage)
	// for x := 0; x < width; x++ {
	// 	for y := 0; y < height; y++ {
	// 		xpixel := C.XGetPixelFunc(ximage, C.int(x), C.int(y))
	// 		img.SetRGBA(x, y, color.RGBA{
	// 			R: uint8((xpixel & ximage.red_mask) >> 16),
	// 			G: uint8((xpixel & ximage.green_mask) >> 8),
	// 			B: uint8((xpixel & ximage.blue_mask)),
	// 		})
	// 	}
	// }
	img.Pix = C.GoBytes(unsafe.Pointer(ximage.data), C.int(width*height*4))
	return img, nil
}
