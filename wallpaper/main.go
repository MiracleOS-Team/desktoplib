package wallpaper

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/xgraphics"
	"github.com/BurntSushi/xgbutil/xprop"
)

func swwwCommand(args []string) (string, error) {
	cmd := exec.Command("swww", args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("error executing swww: %v, stderr: %s, stdout: %s", err, stderr.String(), stdout.String())
	}

	return stdout.String(), nil
}

func startSwww() error {
	cmd := exec.Command("swww-daemon")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	return err
}

func startMPVPaper(options string, monitors string, file string) error {
	cmd := exec.Command("mpvpaper", "-o", options, monitors, file)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	return err
}

func sendMPVPaperCommand(command string) error {
	c, err := net.Dial("unix", "/tmp/mpvpaper-socket")
	if err != nil {
		return err
	}
	defer c.Close()

	_, err = c.Write([]byte(command))
	if err != nil {
		return err
	}

	_, err = c.Write([]byte("\n"))
	return err

}

func StopWallpaper() {
	_, eerr := os.Stat(filepath.Join("/tmp", "mpvpaper-socket"))
	if !os.IsNotExist(eerr) {
		sendMPVPaperCommand("quit")
		os.Remove(filepath.Join("/tmp", "mpvpaper-socket"))
	} else {
		swwwCommand([]string{"clear"})
		swwwCommand([]string{"kill"})
	}
}

func SetVideoWallpaper(file string, displays string, loop bool) error {
	StopWallpaper()

	mpvargs := "no-audio input-ipc-server=/tmp/mpvpaper-socket -f"
	if loop {
		mpvargs += " loop"
	}

	go startMPVPaper(mpvargs, displays, file)
	return nil
}

func setRootPixmapProperties(X *xgbutil.XUtil, root xproto.Window, pixmap xproto.Pixmap) error {
	// Create atoms
	_, err := xprop.Atm(X, "_XROOTPMAP_ID")
	if err != nil {
		return fmt.Errorf("creating _XROOTPMAP_ID atom: %s", err)
	}
	_, err = xprop.Atm(X, "_XSETROOT_ID")
	if err != nil {
		return fmt.Errorf("creating _XSETROOT_ID atom: %s", err)
	}

	// Set the pixmap ID on both atoms
	err = xprop.ChangeProp32(X, root, "_XROOTPMAP_ID", "PIXMAP", uint(pixmap))
	if err != nil {
		return fmt.Errorf("setting _XROOTPMAP_ID: %s", err)
	}
	err = xprop.ChangeProp32(X, root, "_XSETROOT_ID", "PIXMAP", uint(pixmap))
	if err != nil {
		return fmt.Errorf("setting _XSETROOT_ID: %s", err)
	}

	return nil
}

func setWallpaperXorg(file string) error {
	X, err := xgbutil.NewConn()
	if err != nil {
		return err
	}

	go func() {
		defer X.Conn().Close()
		// Load and convert image
		ximg, err := xgraphics.NewFileName(X, file)
		if err != nil {
			log.Fatal("loading image:", err)
		}

		// Resize to screen dimensions
		screen := X.Screen()
		ximg = ximg.Scale(int(screen.WidthInPixels), int(screen.HeightInPixels))

		// Create a Pixmap, draw to it, and paint it to the root
		ximg.CreatePixmap()
		ximg.XDraw()
		ximg.XPaint(X.RootWin())

		// Important: Set _XROOTPMAP_ID so Openbox won't clear it
		err = setRootPixmapProperties(X, X.RootWin(), ximg.Pixmap)
		if err != nil {
			log.Fatal("setting _XROOTPMAP_ID:", err)
		}
	}()

	return nil
}

func SetImageWallpaper(file string, displays string) error {
	StopWallpaper()

	go startSwww()

	args := []string{"img", file}
	if displays != "" {
		args = append(args, displays)
	}

	_, err := swwwCommand(args)
	if err != nil {
		return setWallpaperXorg(file)
	}
	return nil
}
