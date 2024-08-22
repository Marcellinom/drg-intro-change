package main

import (
	"bufio"
	"fmt"
	"golang.org/x/sys/windows/registry"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func getSteamLibraryPaths(reg_key registry.Key) ([]string, error) {
	key, err := registry.OpenKey(reg_key, `SOFTWARE\Valve\Steam`, registry.QUERY_VALUE)
	if err != nil {
		return nil, fmt.Errorf("error open key %w", err)
	}
	steamPath, _, err := key.GetStringValue("SteamPath")
	if err != nil {
		return nil, fmt.Errorf("error get string value %w", err)
	}

	libraryFile := filepath.Join(steamPath, "steamapps", "libraryfolders.vdf")

	file, err := os.Open(libraryFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var libraryPaths []string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "path") {
			libraryPaths = append(libraryPaths, line[10:])
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return libraryPaths, nil
}

func getDrgPath(lib_paths []string) (string, error) {
	var drg_library string
	for _, v := range lib_paths {
		v = strings.ReplaceAll(v, `"`, "")
		drg_library = filepath.Join(v, "steamapps", "common", "Deep Rock Galactic", "FSD", "Content", "Movies")
		_, err := os.Stat(drg_library)
		if err == nil {
			return drg_library, nil
		}
	}
	return "", fmt.Errorf("drg path does not exists: %s", drg_library)
}

func main() {
	fmt.Println("getting steam path")
	paths, err := getSteamLibraryPaths(registry.CURRENT_USER)
	if err != nil {
		paths, err = getSteamLibraryPaths(registry.LOCAL_MACHINE)
		if err != nil {
			fmt.Println("error while getting steam library path ", err)
			fmt.Scanln()
		}
	}
	fmt.Println("getting drg path")
	drg_path, err := getDrgPath(paths)
	if err != nil {
		fmt.Println(err)
		fmt.Scanln()
	}

	intro := filepath.Join(drg_path, `DRG_LogoIntro_720p30.mp4`)
	intro_lower := filepath.Join(drg_path, `DRG_LogoIntro_Lower_Sound_720p30.mp4`)
	output := filepath.Join(drg_path, strconv.FormatInt(time.Now().Unix(), 10)+`_DRG_LogoIntro_720p30.mp4`)
	output_lower := filepath.Join(drg_path, strconv.FormatInt(time.Now().Unix(), 10)+`_DRG_LogoIntro_Lower_Sound_720p30.mp4`)

	fmt.Println("renaming old intro")
	os.Rename(intro, output)
	os.Rename(intro_lower, output_lower)

	fmt.Print("paste a fetchable video link to be downloaded for new DRG intro: ")
	var media string
	_, err = fmt.Scanln(&media)
	if err != nil {
		fmt.Println(err)
		fmt.Scanln()
	}

	fmt.Println("downloading new intro")
	err = downloadFile(
		media,
		intro)
	if err != nil {
		fmt.Println("failed to download media ", err)
		fmt.Scanln()
	}
	err = downloadFile(
		media,
		intro_lower)
	if err != nil {
		fmt.Println("failed to download media ", err)
		fmt.Scanln()
	}
	fmt.Println("success")
	fmt.Scanln()
}

// downloadFile downloads a file from the given URL and saves it to the given path
func downloadFile(url, filepath string) error {
	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Send GET request to the URL
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	// Check if the server response was successful
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	_, err = io.Copy(out, resp.Body)

	return err
}
