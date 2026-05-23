package engines

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"time"
)

type YaraScanResult struct {
	HasMatches bool
	Matches    []YaraMatch
	RawOutput  string
	Duration   time.Duration
	Error      error
}

type YaraMatch struct {
	RuleName string
	FilePath string
}

type YaraScannerDockerEngine interface {
	Check() error
	Scan(ctx context.Context, targetPath string) (*YaraScanResult, error)
}

type yaraScannerEngine struct {
	containerName string
	scanRoot      string
	executor      CommandExecutor
}

func NewYaraScannerDockerEngine(excutor CommandExecutor, containerName, scanRoot string) (YaraScannerDockerEngine, error) {
	if containerName == "" {
		return nil, errors.New("container name is not prepared")
	}

	if scanRoot == "" {
		return nil, errors.New("scan root is not prepared")
	}

	return &yaraScannerEngine{
		containerName: containerName,
		scanRoot:      scanRoot,
		executor:      excutor,
	}, nil
}

// Check for docker.sock and exists the yara image and test rules
func (r *yaraScannerEngine) Check() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// check docker socket for connection to external docker deamon
	if _, err := os.Stat("/var/run/docker.sock"); err != nil {
		return fmt.Errorf("docker socket not accessible: %w", err)
	}

	if err := r.checkContainerIsRunning(ctx); err != nil {
		return fmt.Errorf("docker container check failed: %w", err)
	}

	return nil
}

func (r *yaraScannerEngine) checkContainerIsRunning(ctx context.Context) error {
	fmt.Println("docker", "container", "inspect", r.containerName)
	output, err := r.executor.ExecCommand(ctx, "docker", "container", "inspect", r.containerName)
	if err != nil {
		if strings.Contains(string(output), "No such container") {
			return fmt.Errorf("docker container '%s' does not exist", r.containerName)
		}
		return fmt.Errorf("failed to inspect container: %w, output: %s", err, string(output))
	}
	return nil
}

// Scan target file
func (r *yaraScannerEngine) Scan(ctx context.Context, targetFile string) (*YaraScanResult, error) {
	startTime := time.Now()

	result := &YaraScanResult{
		Matches: []YaraMatch{},
	}
	// Copy a target file to scan root directory for scanning
	dst := path.Join(r.scanRoot, path.Base(targetFile))
	if err := copyFile(targetFile, dst); err != nil {
		return nil, err
	}

	// Run the yara scan command in docker container
	// Command format:
	// docker exec -it yara-service yara /rules/*.yar /path/to/target/file
	args := []string{}
	args = append(args, dst)
	output, err := r.execYaraCommand(ctx, args...)

	result.Duration = time.Since(startTime)
	result.RawOutput = string(output)

	if err != nil {
		result.Error = err
		return result, err
	}

	result.Matches = r.parseYaraOutput(string(output))
	result.HasMatches = len(result.Matches) > 0

	return result, nil
}

func (r *yaraScannerEngine) execYaraCommand(ctx context.Context, args ...string) ([]byte, error) {
	dockerArgs := []string{
		"exec",
	}

	dockerArgs = append(dockerArgs, r.containerName)
	dockerArgs = append(dockerArgs, "sh")
	dockerArgs = append(dockerArgs, "-c")
	dockerArgs = append(dockerArgs, fmt.Sprintf("yara /rules/*.yar %s", strings.Join(args, " ")))
	fmt.Println("yara command : ", append([]string{"docker"}, dockerArgs...))
	return r.executor.ExecCommand(ctx, "docker", dockerArgs...)
}

func (r *yaraScannerEngine) parseYaraOutput(output string) []YaraMatch {
	lines := strings.Split(output, "\n")
	matches := []YaraMatch{}

	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, " ", 2)
		if len(parts) == 2 {
			match := YaraMatch{
				RuleName: parts[0],
				FilePath: parts[1],
			}
			matches = append(matches, match)
		}
	}

	return matches
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	return destFile.Sync()
}
