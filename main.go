package main

import (
	"bufio"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func createLogFile() (*os.File, error) {
	date := time.Now()
	name := date.Format("2006-01-02_15:04:05")
	file, err := os.Create(name)
	if err != nil {
		return nil, err
	}

	return file, nil
}

// Search for a pubspec.yaml file in current directory.
//
// If no file is found, it will look in a parent dir until it reaches the root "/".
func getPubspecFile() (*os.File, error) {
	currentDir := filepath.Dir(os.Getenv("PWD"))
	for {
		if currentDir == "/" {
			break
		}

		tryPath := filepath.Join(currentDir, "pubspec.yaml")
		pubspec, err := os.OpenFile(tryPath, os.O_RDWR, fs.FileMode(0644))
		if err != nil {
			currentDir = filepath.Dir(currentDir)
			continue
		}

		return pubspec, nil
	}

	return nil, fmt.Errorf("Did not found pubscpec file")
}

// Search for the build version in pubspec [file]
//
// Returns [version], [build] and possible [error]
func getVersionFromPubspecFile(file *os.File) (version string, build string, err error) {
	scanner := bufio.NewScanner(file)
	if err := scanner.Err(); err != nil {
		slog.Error(fmt.Sprint(err))
		return "", "", err
	}

	for scanner.Scan() {
		line := scanner.Text()
		if !strings.Contains(line, "version: ") {
			continue
		}

		versionAndBuild := strings.Replace(line, "version: ", "", 1)
		arr := strings.Split(versionAndBuild, "+")
		if len(arr) != 2 {
			err = fmt.Errorf("The build version found in pubspec file is invalid: %s", versionAndBuild)
			slog.Error(fmt.Sprint(err))
			return "", "", err
		}
		return arr[0], arr[1], nil
	}

	return "", "", fmt.Errorf("Could not get the version from pubspec file.")
}

func buildAndroidBundle() error {
	buildCmd := exec.Command("flutter", "build", "appbundle")
	// TODO: implement a writer that uses slog to print buildCmd stdout & stderr
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr
	return buildCmd.Run()
}

// Will print a log with [err] and exit the program with [code] using [os.Exit].
//
// Given [os.Exit] constraints, the status code should be in the range [1, 125].
func logAndExit(err error, code int) {
	slog.Error(fmt.Sprint(err))
	os.Exit(code)
}

func main() {
	// TODO: set new pubspec version
	pubspec, err := getPubspecFile()
	if err != nil {
		logAndExit(err, 1)
	}
	defer pubspec.Close()

	// get current version in pubspec
	var version, build string
	version, build, err = getVersionFromPubspecFile(pubspec)
	if err != nil {
		logAndExit(err, 1)
	}
	fmt.Println(version)
	fmt.Println(build)

	// build android appbundle
	err = buildAndroidBundle()
	if err != nil {
		logAndExit(err, 1)
	}
}
