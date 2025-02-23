package wallpaper

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
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

func SetImageWallpaper(file string, displays string) error {
	StopWallpaper()

	go startSwww()

	args := []string{"img", file}
	if displays != "" {
		args = append(args, displays)
	}

	_, err := swwwCommand(args)
	return err
}
