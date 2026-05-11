package engines

import (
	"context"
	"fmt"
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
	image     string
	rulesRoot string
	executor  CommandExecutor
}

func NewYaraScannerDockerEngine(excutor CommandExecutor, image, rulesRoot string) (YaraScannerDockerEngine, error) {

	// check docker socket for connection to external docker deamon
	if _, err := os.Stat("/var/run/docker.sock"); err != nil {
		return nil, fmt.Errorf("docker socket not accessible: %w", err)
	}

	if !strings.HasSuffix(rulesRoot, "/*.yar") {
		rulesRoot = path.Join(rulesRoot, "/*.yar")
	}

	return &yaraScannerEngine{
		image:     image,
		rulesRoot: rulesRoot,
	}, nil
}

// Check for exists the yara image and test rules
func (r *yaraScannerEngine) Check() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := r.checkImageExists(ctx); err != nil {
		return fmt.Errorf("docker image check failed: %w", err)
	}

	if err := r.checkRulesRoot(ctx); err != nil {
		return fmt.Errorf("rules root check failed: %w", err)
	}

	return nil
}

func (r *yaraScannerEngine) checkImageExists(ctx context.Context) error {
	output, err := r.executor.ExecCommand(ctx, "docker", "image", "inspect", r.image)
	if err != nil {
		if strings.Contains(string(output), "No such image") {
			return fmt.Errorf("docker image '%s' does not exist", r.image)
		}
		return fmt.Errorf("failed to inspect image: %w, output: %s", err, string(output))
	}
	return nil
}

func (r *yaraScannerEngine) checkRulesRoot(ctx context.Context) error {
	_, err := r.execYaraCommand(ctx, r.rulesRoot, "/dev/null")
	if err != nil {
		return fmt.Errorf("rules root directory '%s' not accessible: %w", r.rulesRoot, err)
	}
	return nil
}

// Scan target file
func (r *yaraScannerEngine) Scan(ctx context.Context, targetFile string) (*YaraScanResult, error) {
	startTime := time.Now()

	result := &YaraScanResult{
		Matches: []YaraMatch{},
	}

	args := []string{}
	args = append(args, r.rulesRoot)
	args = append(args, targetFile)
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
		"run",
		"--rm",
		"-v", fmt.Sprintf("%s:%s:ro", r.rulesRoot, r.rulesRoot),
	}

	dockerArgs = append(dockerArgs, r.image)
	dockerArgs = append(dockerArgs, args...)

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

func (r *yaraScannerEngine) GetImage() string {
	return r.image
}
