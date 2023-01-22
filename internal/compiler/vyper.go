// Copyright 2019 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

// Package compiler wraps the Solidity and Vyper compiler executables (solc; vyper).
package compiler

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strconv"
	"strings"
)

// Vyper contains information about the vyper compiler.
type Vyper struct {
	Path, Version, FullVersion string
	Major, Minor, Patch        int
}

func (s *Vyper) makeArgs() []string {
	p := []string{
		"-f", "combined_json",
	}
	return p
}

// VyperVersion runs vyper and parses its version output.
func VyperVersion(vyper string) (*Vyper, error) {
	if vyper == "" {
		vyper = "vyper"
	}
	var out bytes.Buffer
	cmd := exec.Command(vyper, "--version")
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return nil, err
	}
	matches := versionRegexp.FindStringSubmatch(out.String())
	if len(matches) != 4 {
		return nil, fmt.Errorf("can't parse vyper version %q", out.String())
	}
	s := &Vyper{Path: cmd.Path, FullVersion: out.String(), Version: matches[0]}
	if s.Major, err = strconv.Atoi(matches[1]); err != nil {
		return nil, err
	}
	if s.Minor, err = strconv.Atoi(matches[2]); err != nil {
		return nil, err
	}
	if s.Patch, err = strconv.Atoi(matches[3]); err != nil {
		return nil, err
	}
	return s, nil
}

// CompileVyper compiles all given Vyper source files.
func CompileVyper(vyper string, sourcefiles ...string) (map[string]*Contract, error) {
	if len(sourcefiles) == 0 {
		return nil, errors.New("vyper: no source files")
	}
	source, err := slurpFiles(sourcefiles)
	if err != nil {
		return nil, err
	}
	s, err := VyperVersion(vyper)
	if err != nil {
		return nil, err
	}
	args := s.makeArgs()
	cmd := exec.Command(s.Path, append(args, sourcefiles...)...)
	return s.run(cmd, source)
}

func (s *Vyper) run(cmd *exec.Cmd, source string) (map[string]*Contract, error) {
	var stderr, stdout bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("vyper: %v\n%s", err, stderr.Bytes())
	}

	return ParseVyperJSON(stdout.Bytes(), source, s.Version, s.Version, strings.Join(s.makeArgs(), " "))
}

// ParseVyperJSON takes the direct output of a vyper --f combined_json run and
// parses it into a map of string contract name to Contract structs. The
// provided source, language and compiler version, and compiler options are all
// passed through into the Contract structs.
//
// The vyper output is expected to contain ABI and source mapping.
//
// Returns an error if the JSON is malformed or missing data, or if the JSON
// embedded within the JSON is malformed.
func ParseVyperJSON(combinedJSON []byte, source string, languageVersion string, compilerVersion string, compilerOptions string) (map[string]*Contract, error) {
	var output map[string]interface{}
	if err := json.Unmarshal(combinedJSON, &output); err != nil {
		return nil, err
	}

	// Compilation succeeded, assemble and return the contracts.
	contracts := make(map[string]*Contract)
	for name, info := range output {
		// Parse the individual compilation results.
		if name == "version" {
			continue
		}
		c := info.(map[string]interface{})

		contracts[name] = &Contract{
			Code:        c["bytecode"].(string),
			RuntimeCode: c["bytecode_runtime"].(string),
			Info: ContractInfo{
				Source:          source,
				Language:        "Vyper",
				LanguageVersion: languageVersion,
				CompilerVersion: compilerVersion,
				CompilerOptions: compilerOptions,
				SrcMap:          c["source_map"],
				SrcMapRuntime:   "",
				AbiDefinition:   c["abi"],
				UserDoc:         "",
				DeveloperDoc:    "",
				Metadata:        "",
			},
		}
	}
	return contracts, nil
}

type VyperCompiler struct {
	version string
	path    string
}

func NewVyperCompiler(version string) (Compiler, error) {
	compilerDir := path.Join(os.TempDir(), "vyper", version)

	var compilerPath string
	if runtime.GOOS == "windows" {
		compilerPath = path.Join(compilerDir, "vyper.exe")
	} else {
		compilerPath = path.Join(compilerDir, "vyper")
	}

	if _, err := os.Stat(compilerPath); err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to check compiler: %w", err)
		}
		err = downloadVyperCompiler(version, compilerDir)
		if err != nil {
			return nil, fmt.Errorf("failed to download compiler: %w", err)
		}
	}

	return &VyperCompiler{
		version: version,
		path:    compilerPath,
	}, nil
}

func downloadVyperCompiler(version string, dest string) error {
	err := os.MkdirAll(dest, os.FileMode(0755))
	if err != nil {
		return fmt.Errorf("failed to create parent dir: %w", err)
	}

	resp, err := http.Get(fmt.Sprintf(`https://api.github.com/repos/vyperlang/vyper/releases/tags/v%s`, version))
	if err != nil {
		return fmt.Errorf("failed to get release manifest: %w", err)
	}

	var release struct {
		Assets []struct {
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
		} `json:"assets"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return fmt.Errorf("failed to decode manifest: %w", err)
	}

	var url string
	var name string

	for _, asset := range release.Assets {
		if runtime.GOOS == "linux" && strings.Contains(asset.Name, "linux") {
			url = asset.BrowserDownloadURL
			name = "vyper"
		} else if runtime.GOOS == "windows" && strings.Contains(asset.Name, "windows") {
			url = asset.BrowserDownloadURL
			name = "vyper.exe"
		}
	}

	if url == "" {
		return fmt.Errorf("unsupported os %s", runtime.GOOS)
	}

	resp, err = http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download solidity v%s: %w", version, err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read body: %w", err)
	}

	err = ioutil.WriteFile(path.Join(dest, name), body, os.FileMode(0755))
	if err != nil {
		return fmt.Errorf("failed to write compiler: %w", err)
	}

	return nil
}

func (c *VyperCompiler) CompileFromString(src string) (map[string]*Contract, error) {
	s, err := VyperVersion(c.path)
	if err != nil {
		return nil, err
	}
	args := s.makeArgs()
	cmd := exec.Command(s.Path, append(args, "/dev/stdin")...)
	return c.run(s, cmd, src)
}

func (c *VyperCompiler) run(s *Vyper, cmd *exec.Cmd, source string) (map[string]*Contract, error) {
	var stderr, stdout bytes.Buffer
	cmd.Stdin = strings.NewReader(source)
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("vyper: %v\n%s", err, stderr.Bytes())
	}

	return ParseVyperJSON(stdout.Bytes(), source, s.Version, s.Version, strings.Join(s.makeArgs(), " "))
}
