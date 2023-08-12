package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/google/go-github/v33/github"
	"golang.org/x/oauth2"
)

const (
	oneAPIURL      = "https://github.com/songquanpeng/one-api/releases/download/v0.5.2/one-api"
	oneAPIFileName = "one-api"
	owner          = "nextyancc"
	repo           = "one-api-files"
	oneAPIDBName   = "one-api.db"
)

func main() {
	if err := runOneAPI(); err != nil {
		fmt.Println("Error:", err)
	}
}

func runOneAPI() error {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		log.Fatal("Unauthorized: No token present")
	}
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(context.Background(), ts)

	client := github.NewClient(tc)
	downloadFile(client, owner, repo, oneAPIFileName)
	fmt.Printf("下载完成%s..\n", oneAPIFileName)
	downloadFile(client, owner, repo, oneAPIDBName)
	fmt.Printf("下载完成%s..\n", oneAPIDBName)
	// if err := downloadOneAPI(oneAPIFileName, oneAPIURL); err != nil {
	// 	return fmt.Errorf("download error: %w", err)
	// }

	if err := os.Chmod(oneAPIFileName, 0755); err != nil {
		return fmt.Errorf("permission error: %w", err)
	}

	cmd := exec.Command("./" + oneAPIFileName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("execution error: %w", err)
	}

	return nil
}

func downloadOneAPI(filepath string, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("HTTP GET error: %w", err)
	}
	defer resp.Body.Close()

	out, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("file creation error: %w", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		return fmt.Errorf("file copy error: %w", err)
	}

	return nil
}

func uploadFile(client *github.Client, owner, repo, path string, content []byte) {
	ctx := context.Background()
	message := "备份时间：" + time.Now().Format("2006-01-02 15:04:05")
	branch := "main"
	opts := &github.RepositoryContentFileOptions{
		Message: &message,
		Content: content,
		Branch:  &branch,
	}
	_, _, err := client.Repositories.CreateFile(ctx, owner, repo, path, opts)
	if err != nil {
		log.Fatal(err)
	}
}

func downloadFile(client *github.Client, owner, repo, path string) {
	ctx := context.Background()
	fileContent, _, _, err := client.Repositories.GetContents(ctx, owner, repo, path, &github.RepositoryContentGetOptions{})
	if err != nil {
		log.Fatal(err)
	}

	if fileContent.GetEncoding() != "base64" {
		resp, err := http.Get(fileContent.GetDownloadURL())
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		err = os.WriteFile(path, body, 0644)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		content, err := fileContent.GetContent()
		if err != nil {
			log.Fatal(err)
		}
		err = os.WriteFile(path, []byte(content), 0644)
		if err != nil {
			log.Fatal(err)
		}
	}
}
