/*
	Grabs a screenshot of the root window.
	
	Usage	: ./scr_tool <display> <output file>
	Example	: ./scr_tool :0 /path/to/output.png

	Author: S Bozdag <selcuk.bozdag@gmail.com>
*/

#include <assert.h>
#include <stdio.h>
#include <cairo.h>
#include <cairo-xlib.h>
#include <X11/Xlib.h>
#include <X11/extensions/Xinerama.h>
#include <X11/extensions/XShm.h>
#include <X11/Xutil.h>

void printScreens()
{
    Display *display = XOpenDisplay(NULL);
    if (display == NULL)
    {
        return;
    }
    int nmonitors = 0;
    XineramaScreenInfo *screen = XineramaQueryScreens(display, &nmonitors);
    if (screen == NULL)
    {
        XCloseDisplay(display);
        return;
    }
    for (int i = 0; i < nmonitors; i++)
    {

        printf("%d: %d %dx%d %d %d\n", i, screen[i].screen_number, screen[i].width, screen[i].height, screen[i].x_org, screen[i].y_org);
    }
    XFree(screen);
    XCloseDisplay(display);
}

XineramaScreenInfo *getScreen(int num)
{
    Display *display = XOpenDisplay(NULL);
    if (display == NULL)
    {
        return NULL;
    }
    int nmonitors = 0;
    XineramaScreenInfo *screen = XineramaQueryScreens(display, &nmonitors);
    if (screen == NULL)
    {
        XCloseDisplay(display);
        return NULL;
    }
    if (num >= nmonitors)
    {
        XFree(screen);
        XCloseDisplay(display);
        return NULL;
    }
    XCloseDisplay(display);
    return &screen[num];
}

int main(int argc, char **argv)
{
    printScreens();
    XineramaScreenInfo *screenInfo = getScreen(0);
    Display *display = XOpenDisplay(NULL);
    if (display == NULL)
    {
        return 1;
    }
    int scr = XDefaultScreen(display);
    XShmSegmentInfo shminfo;
    XImage *XImage_;
    XImage_ = XGetImage(display,
                        DefaultRootWindow(display),
                        screenInfo->x_org,
                        screenInfo->y_org,
                        screenInfo->width,
                        screenInfo->height,
                        AllPlanes, ZPixmap);
    if (XImage_ == NULL)
    {
        printf("image is null\n");
        return 1;
    }
    printf("height: %d\n", XImage_->height);
    printf("stride: %d\n", XImage_->bytes_per_line);
    cairo_surface_t *surface =
        cairo_image_surface_create_for_data(
            XImage_->data,
            XImage_->depth == 24 ? CAIRO_FORMAT_RGB24 : CAIRO_FORMAT_ARGB32,
            XImage_->width,
            XImage_->height,
            XImage_->bytes_per_line);
    if (surface == NULL)
    {
        printf("XXXXXXX\n");
        return 1;
    }
    cairo_status_t sta = cairo_surface_write_to_png(surface, "./abrakadabra.png");
    if (sta != CAIRO_STATUS_SUCCESS)
    {
        printf("FAIL PNG %s\n", cairo_status_to_string(sta));
        return 1;
    }
    cairo_surface_destroy(surface);
    return 0;
}