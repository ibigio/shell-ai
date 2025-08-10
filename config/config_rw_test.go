package config

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	types "q/types"

	"gopkg.in/yaml.v2"
)

// helper to point config paths to a temp dir and restore after
func withTempConfigPaths(t *testing.T) (tmp string, restore func()) {
	t.Helper()
	tmpDir := t.TempDir()
	oldCfg := configFilePath
	oldBak := backupConfigFilePath
	// Point HOME to tmp so FullFilePath writes under tmp
	t.Setenv("HOME", tmpDir)
	configFilePath = ".shell-ai/config.yaml"
	backupConfigFilePath = ".shell-ai/.backup-config.yaml"
	return tmpDir, func() {
		configFilePath = oldCfg
		backupConfigFilePath = oldBak
	}
}

func readYAML(t *testing.T, path string, out interface{}) {
	t.Helper()
	b, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if err := yaml.Unmarshal(b, out); err != nil {
		t.Fatalf("yaml unmarshal %s: %v", path, err)
	}
}

func TestLoadAppConfig_CreatesDefaultAndRespectsOverride(t *testing.T) {
	_, restore := withTempConfigPaths(t)
	defer restore()

	t.Setenv("OPENAI_MODEL_OVERRIDE", "override-model")

	// ensure file does not exist
	abs, _ := FullFilePath(configFilePath)
	if _, err := os.Stat(abs); !os.IsNotExist(err) {
		t.Fatalf("expected no config file at start")
	}

	cfg, err := LoadAppConfig()
	if err != nil {
		t.Fatalf("LoadAppConfig: %v", err)
	}
	if cfg.Preferences.DefaultModel != "override-model" {
		t.Fatalf("override not applied, got %q", cfg.Preferences.DefaultModel)
	}
	if _, err := os.Stat(abs); err != nil {
		t.Fatalf("expected config file to be created: %v", err)
	}
}

func TestSaveAppConfig_WritesMainAndBackup(t *testing.T) {
	_, restore := withTempConfigPaths(t)
	defer restore()

	app := AppConfig{
		Preferences: types.Preferences{DefaultModel: "gpt-5"},
		Models: []types.ModelConfig{
			{
				ModelName: "gpt-5",
				Endpoint:  "https://api.openai.com/v1/chat/completions",
				Auth:      "SECRET",
				OrgID:     "ORG",
				Prompt:    []types.Message{{Role: "system", Content: "hi"}},
			},
		},
		Version: "1",
	}

	if err := SaveAppConfig(app); err != nil {
		t.Fatalf("SaveAppConfig: %v", err)
	}

	// main file
	abs, _ := FullFilePath(configFilePath)
	var gotMain AppConfig
	readYAML(t, abs, &gotMain)
	if gotMain.Preferences.DefaultModel != "gpt-5" || len(gotMain.Models) != 1 {
		t.Fatalf("unexpected main config: %+v", gotMain)
	}

	// backup file
	bak, _ := FullFilePath(backupConfigFilePath)
	var gotBackup AppConfig
	readYAML(t, bak, &gotBackup)
	if gotBackup.Preferences.DefaultModel != "gpt-5" || len(gotBackup.Models) != 1 {
		t.Fatalf("unexpected backup config: %+v", gotBackup)
	}
}

func TestRevertAppConfigToBackupRestores(t *testing.T) {
	_, restore := withTempConfigPaths(t)
	defer restore()

	// write an initial config and backup via SaveAppConfig
	base := AppConfig{
		Preferences: types.Preferences{DefaultModel: "base"},
		Models:      []types.ModelConfig{},
		Version:     "1",
	}
	if err := SaveAppConfig(base); err != nil {
		t.Fatalf("SaveAppConfig (base): %v", err)
	}

	// now corrupt/replace main config
	abs, _ := FullFilePath(configFilePath)
	if err := ioutil.WriteFile(abs, []byte("preferences:\n  default_model: changed\n"), 0644); err != nil {
		t.Fatalf("write corrupt main: %v", err)
	}

	if err := RevertAppConfigToBackup(); err != nil {
		t.Fatalf("RevertAppConfigToBackup: %v", err)
	}

	var got AppConfig
	readYAML(t, abs, &got)
	if got.Preferences.DefaultModel != "base" {
		t.Fatalf("revert did not restore backup, got %q", got.Preferences.DefaultModel)
	}
}

func TestResetAppConfigToDefault_WritesEmbeddedAndBackup(t *testing.T) {
	_, restore := withTempConfigPaths(t)
	defer restore()

	// first call should create defaults
	if err := ResetAppConfigToDefault(); err != nil {
		t.Fatalf("ResetAppConfigToDefault: %v", err)
	}

	abs, _ := FullFilePath(configFilePath)
	bak, _ := FullFilePath(backupConfigFilePath)
	var gotMain, gotBackup AppConfig
	readYAML(t, abs, &gotMain)
	readYAML(t, bak, &gotBackup)
	if gotMain.Version == "" {
		t.Fatalf("expected embedded defaults to set version, got empty")
	}
	if gotBackup.Version != gotMain.Version {
		t.Fatalf("backup version mismatch: %q vs %q", gotBackup.Version, gotMain.Version)
	}
}

func TestEnvOverride_AppliesOnlyOnCreate_NotOnLoad(t *testing.T) {
	_, restore := withTempConfigPaths(t)
	defer restore()

	// Create default with override
	t.Setenv("OPENAI_MODEL_OVERRIDE", "override-model")
	cfg, err := LoadAppConfig()
	if err != nil {
		t.Fatalf("LoadAppConfig on create: %v", err)
	}
	if cfg.Preferences.DefaultModel != "override-model" {
		t.Fatalf("override not applied on create")
	}

	// Save a different value
	cfg.Preferences.DefaultModel = "saved-model"
	if err := SaveAppConfig(cfg); err != nil {
		t.Fatalf("SaveAppConfig: %v", err)
	}

	// Change the env var and ensure load does NOT override
	t.Setenv("OPENAI_MODEL_OVERRIDE", "different")
	loaded, err := LoadAppConfig()
	if err != nil {
		t.Fatalf("LoadAppConfig existing: %v", err)
	}
	if loaded.Preferences.DefaultModel != "saved-model" {
		t.Fatalf("env override should not change existing file; got %q", loaded.Preferences.DefaultModel)
	}
}

func TestLoadAppConfig_InvalidYAML_ReturnsError(t *testing.T) {
	_, restore := withTempConfigPaths(t)
	defer restore()

	// Write invalid YAML
	abs, _ := FullFilePath(configFilePath)
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := ioutil.WriteFile(abs, []byte("preferences: [this is not valid: yaml"), 0644); err != nil {
		t.Fatalf("write invalid yaml: %v", err)
	}
	_, err := LoadAppConfig()
	if err == nil {
		t.Fatalf("expected error for invalid YAML, got nil")
	}
}

func TestRevertAppConfigToBackup_NoBackup_ReturnsError(t *testing.T) {
	_, restore := withTempConfigPaths(t)
	defer restore()

	// Ensure no backup exists, but main exists
	cfg := AppConfig{Preferences: types.Preferences{DefaultModel: "m"}}
	if err := SaveAppConfig(cfg); err != nil {
		t.Fatalf("SaveAppConfig: %v", err)
	}
	// Delete backup
	bak, _ := FullFilePath(backupConfigFilePath)
	_ = os.Remove(bak)

	if err := RevertAppConfigToBackup(); err == nil {
		t.Fatalf("expected error when backup missing, got nil")
	}
}

func TestFullFilePath_ResolvesUnderHome(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	p, err := FullFilePath(".shell-ai/foo.yaml")
	if err != nil {
		t.Fatalf("FullFilePath error: %v", err)
	}
	if p != filepath.Join(tmp, ".shell-ai", "foo.yaml") {
		t.Fatalf("unexpected path: %s", p)
	}
}

func TestSaveBackupConfig_ErrorWhenPathIsDirectory(t *testing.T) {
	_, restore := withTempConfigPaths(t)
	defer restore()

	// Create a directory where backup file should be
	bak, _ := FullFilePath(backupConfigFilePath)
	if err := os.MkdirAll(bak, 0o755); err != nil { // mkdir path with filename to force EISDIR on write
		t.Fatalf("mkdir backup path: %v", err)
	}
	err := SaveBackupConfig(AppConfig{})
	if err == nil {
		t.Fatalf("expected error when backup path is a directory")
	}
}
