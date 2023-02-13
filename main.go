package main

import (
	_ "embed"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"io/ioutil"
	"time"

	"github.com/getlantern/systray"
	"github.com/ncruces/zenity"
)

//go:embed triforce.png
var triforce []byte

func main() {
	executablePath, err := os.Executable()
	if err != nil {
		return
	}

	executableDirectory := filepath.Dir(executablePath)

	homedir, err := os.UserHomeDir()
	if err != nil {
		return
	}

	binaryName := "hyrule-cli-apple-silicon"
	if runtime.GOARCH == "x86" {
		binaryName = "hyrule-cli-intel-mac"
	}

	watchPath := filepath.Join(homedir, "Library/Application Support/Hyrule Desktop")
	cmd := exec.Command(filepath.Join(executableDirectory, binaryName), "--copy", "--watch", watchPath)

	env := os.Environ()
	env = append(env, "LANG=en_US.UTF-8")

	cmd.Env = env

	defer func() {
		cmd.Process.Kill()
	}()

	cmd.Start()

	systray.Run(func() {
		systray.SetIcon(triforce)
		configFileButton := systray.AddMenuItem("Import config", "")
		// setWatchedButton := systray.AddMenuItem("Set Watched Folder", "")
		systray.AddSeparator()
		quitButton := systray.AddMenuItem("Quit", "Quit the whole app")

		go func() {
			for {
				select {
				case <-configFileButton.ClickedCh:
					selectedFilePath, err := zenity.SelectFile(
						zenity.Filename(""),
						zenity.FileFilters{
							{Name: "ShareX configs", Patterns: []string{"*.json", "*.sxcu"}, CaseFold: false},
						})

					if err != nil {
						print(err)
						break
					}

					fileContents, err := ioutil.ReadFile(selectedFilePath)

					if err != nil {
						print(err)
						break
					}

					homedir, err := os.UserHomeDir()

					if err != nil {
						print(err)
						break
					}

					configFileDirectory := filepath.Join(homedir, "Library/Application Support/hyrule-cli/")
					_, err = os.Stat(configFileDirectory)
					if err != nil {
						os.Mkdir(configFileDirectory, 0755)
					}

					err = ioutil.WriteFile(filepath.Join(configFileDirectory, "hyrule.json"), fileContents, 0644)
					if err != nil {
						break
					}

					executablePath, err := os.Executable()

					if err != nil {
						break
					}

					cmd.Process.Kill()

					cmd := exec.Command(executablePath)
					cmd.Start()

					// no idea why this is necessary
					time.Sleep(300 * time.Millisecond)
					os.Exit(0)

				// case <-setWatchedButton.ClickedCh:
				// 	println("lol")

				case <-quitButton.ClickedCh:
					systray.Quit()
				}
			}
		}()
	}, func() {
		cmd.Process.Kill()
	})
}
