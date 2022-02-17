package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/alessio/shellescape"
)

type ScriptGenerator struct {
	out                                              io.WriteCloser
	breakPerChar, breakPerLineBegin, breakPerLineEnd time.Duration
}

func NewScriptGeneratorWithTempFile() (*ScriptGenerator, string, error) {
	scriptFile, err := ioutil.TempFile("", "")
	if err != nil {
		return nil, "", err
	}
	scriptFilePath := scriptFile.Name()

	gen := &ScriptGenerator{
		out:               scriptFile,
		breakPerChar:      100 * time.Millisecond,
		breakPerLineBegin: 500 * time.Millisecond,
		breakPerLineEnd:   500 * time.Millisecond,
	}
	return gen, scriptFilePath, nil
}

func (gen *ScriptGenerator) pr(str string) {
	fmt.Fprintf(gen.out, "echo -n %s\n", shellescape.Quote(str))
}

func (gen *ScriptGenerator) newline() {
	fmt.Fprintf(gen.out, "echo\n")
}

func (gen *ScriptGenerator) sleep(dur time.Duration) {
	fmt.Fprintf(gen.out, "sleep %f\n", dur.Seconds())
}

func (gen *ScriptGenerator) displayLine(line string) {
	gen.pr("$ ")
	gen.sleep(gen.breakPerLineBegin)
	for _, char := range line {
		gen.pr(string(char))
		gen.sleep(gen.breakPerChar)
	}
	gen.sleep(gen.breakPerLineEnd)
	gen.newline()
}

func (gen *ScriptGenerator) execLine(line string) {
	fmt.Fprintf(gen.out, "%s\n", line)
}

func (gen *ScriptGenerator) Close() {
	gen.out.Close()
}

func (gen *ScriptGenerator) ReadLibretto(r io.Reader) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()

		gen.displayLine(line)
		gen.execLine(line)
	}
}

func run(librettoPath string) error {
	file, err := os.Open(librettoPath)
	if err != nil {
		return err
	}
	defer file.Close()

	gen, tmpFilePath, err := NewScriptGeneratorWithTempFile()
	if err != nil {
		return err
	}
	defer os.Remove(tmpFilePath)
	gen.ReadLibretto(file)
	gen.Close()

	cmd := exec.Command("bash", tmpFilePath)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Run()

	return nil
}

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("Usage: %s LIBRETTO-PATH", os.Args[0])
	}
	librettoPath := os.Args[1]

	err := run(librettoPath)
	if err != nil {
		log.Fatalf("ERROR: %v", err)
	}
}
