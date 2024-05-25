package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	_ "embed"
	"encoding/base32"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"
)

//go:embed prompt.txt
var prompt string

// log without timestamp
func init() { log.SetFlags(0) }

func main() {
	debug := flag.Bool("debug", false, "use debug to print the generated awk script")
	dry := flag.Bool("dry", false, "use dry to print the generated awk script without executing it")
	noCache := flag.Bool("no-cache", false, "use no-cache if you don't want to use the cached script for the task")
	sampleLines := flag.Int("lines", 10, "use lines to specify the amount of sample lines the model uses")

	flag.Parse()

	args := flag.Args()
	if len(args) != 1 {
		log.Fatal("You should enter a task that you want to do")
	}

	task := args[0]
	filename, inCache := findCachedScriptFile(task)

	useCache := inCache && !*noCache
	printScript := *dry || *debug

	if printScript && useCache {
		data, err := os.ReadFile(filename)
		if err != nil {
			log.Fatal("Failed reading cached file", err)
		}
		log.Println(string(data))
	}

	rd := io.Reader(os.Stdin)

	if !useCache {
		scanner := bufio.NewScanner(os.Stdin)
		lines := []string{}
		lineCount := 0
		for scanner.Scan() && lineCount < *sampleLines {
			lines = append(lines, scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			log.Fatal("Failed reading from stdin", err)
		}

		sampleData := strings.Join(lines, "\n")
		script, err := getAWKScript(task, sampleData)
		if err != nil {
			log.Fatal("Failed getting AWK script", err)
		}
		if printScript {
			log.Println(script)
		}

		func() { // used to always close file
			f, err := os.Create(filename)
			if err != nil {
				log.Fatal("Cannot create temp file", err)
			}
			defer f.Close()

			_, err = f.Write([]byte(script))
			if err != nil {
				log.Fatal("Cannot write script to temp file", err)
			}
		}()

		// first read sample data and then remaining input
		rd = io.MultiReader(bytes.NewBufferString(sampleData), os.Stdin)
	}

	// if a dry run, don't execute the script
	if *dry {
		return
	}

	// execute script
	cmd := exec.Command("gawk", "-f", filename)

	cmd.Stdin = rd
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		log.Fatal("Cannot execute command", err)
	}
}

// get awk script gets an awk script from openAI with the supplied task (prompt) and sample data
func getAWKScript(task, sampleData string) (string, error) {
	req := ChatRequest{
		Model: ModelGPT35Turbo,
		Messages: []Message{
			{Role: RoleSystem, Content: prompt},
			{Role: RoleUser, Content: fmt.Sprintf("Data:\n%s\n\nTask:\n%s", string(sampleData), task)},
		},
	}

	resp := ChatResponse{}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := openAIRequest(ctx, http.MethodPost, "/chat/completions", &req, &resp)
	if err != nil {
		return "", fmt.Errorf("openAI request failed: %w", err)
	}

	if resp.Error != nil {
		return "", fmt.Errorf("chat returned an error: %s", resp.Error.Message)
	}

	script := resp.Choices[0].Message.Content

	// remove script tags if present
	script = strings.TrimPrefix(script, "```")
	script = strings.TrimPrefix(script, "awk")
	script = strings.TrimPrefix(script, "\n")
	script = strings.TrimSuffix(script, "```")

	return script, nil
}

// findCachedScriptFile checks if there is a cached script for the task
func findCachedScriptFile(task string) (string, bool) {
	sum := sha256.Sum256([]byte(task))
	encoder := base32.StdEncoding.WithPadding(base32.NoPadding)
	hashedTask := encoder.EncodeToString(sum[:])
	filename := "awkai-" + hashedTask + ".awk"

	fullFileName := path.Join(os.TempDir(), filename)
	_, err := os.Stat(fullFileName)

	return fullFileName, err == nil
}
