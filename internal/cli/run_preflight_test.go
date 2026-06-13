package cli

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/gentleman-programming/gentle-ai/internal/installcmd"
	"github.com/gentleman-programming/gentle-ai/internal/model"
	"github.com/gentleman-programming/gentle-ai/internal/system"
)

func TestCheckDependenciesStepFailsWhenKimiUVMissing(t *testing.T) {
	restore := installcmd.OverrideLookPath(func(file string) (string, error) {
		if file == "uv" {
			return "", errNotFound{}
		}
		return "/usr/bin/" + file, nil
	})
	t.Cleanup(restore)

	step := checkDependenciesStep{
		id:      "prepare:check-dependencies",
		profile: system.PlatformProfile{OS: "darwin", PackageManager: "brew", Supported: true},
		selection: model.Selection{
			Agents: []model.AgentID{model.AgentKimi},
		},
	}

	err := step.Run()
	if err == nil {
		t.Fatal("checkDependenciesStep.Run() expected error for missing uv when Kimi is selected")
	}

	if !strings.Contains(err.Error(), "Kimi") || !strings.Contains(err.Error(), "uv") {
		t.Fatalf("checkDependenciesStep.Run() error = %q, expected Kimi uv remediation", err.Error())
	}
}

func TestCheckDependenciesStepDoesNotRequireUVForOtherAgents(t *testing.T) {
	restore := installcmd.OverrideLookPath(func(file string) (string, error) {
		return "", errNotFound{}
	})
	t.Cleanup(restore)

	step := checkDependenciesStep{
		id:      "prepare:check-dependencies",
		profile: system.PlatformProfile{OS: "darwin", PackageManager: "brew", Supported: true},
		selection: model.Selection{
			Agents: []model.AgentID{model.AgentClaudeCode},
		},
	}

	if err := step.Run(); err != nil {
		t.Fatalf("checkDependenciesStep.Run() unexpected error = %v", err)
	}
}

func TestCheckDependenciesStepBoundsDependencyDetection(t *testing.T) {
	origDetectDependencies := detectDependencies
	origTimeout := checkDependenciesTimeout
	t.Cleanup(func() {
		detectDependencies = origDetectDependencies
		checkDependenciesTimeout = origTimeout
	})

	checkDependenciesTimeout = 10 * time.Millisecond
	detectDependencies = func(ctx context.Context, _ system.PlatformProfile) system.DependencyReport {
		<-ctx.Done()
		return system.DependencyReport{}
	}

	step := checkDependenciesStep{
		id:      "prepare:check-dependencies",
		profile: system.PlatformProfile{OS: "android", PackageManager: "pkg", Supported: true},
		selection: model.Selection{
			Agents: []model.AgentID{model.AgentOpenCode},
		},
	}

	start := time.Now()
	if err := step.Run(); err != nil {
		t.Fatalf("checkDependenciesStep.Run() unexpected error = %v", err)
	}
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Fatalf("checkDependenciesStep.Run() took %s, expected bounded dependency detection", elapsed)
	}
}

type errNotFound struct{}

func (errNotFound) Error() string { return "not found" }
