package cmd

import (
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"os"
	"path"
	"runtime"
	"testing"
)

func TestInitConfigWithEnv(t *testing.T) {
	setProjectDir()

	expected := "test"
	t.Setenv("APP_LOG_LEVEL", expected)

	initConfig()

	assert.Equal(t, expected, cfg.App.LogLevel)
}

func TestInitConfigFromFlag(t *testing.T) {
	setProjectDir()

	cfgFile = "./testdata/test.yaml"

	initConfig()

	assert.Equal(t, cfgFile, viper.ConfigFileUsed())
}

func setProjectDir() {
	_, filename, _, _ := runtime.Caller(0)
	dir := path.Join(path.Dir(filename), "../")
	_ = os.Chdir(dir)

	viper.Reset()
}

func TestExecuteCommand_InvalidFlag(t *testing.T) {
	setProjectDir()
	cfgFile = "./testdata/test.yaml"

	rootCmd.SetArgs([]string{"--invalid-flag", "test"})

	err := rootCmd.Execute()

	assert.EqualError(t, err, "unknown flag: --invalid-flag")
}

func TestExecuteCommand_EmptyProject(t *testing.T) {
	setProjectDir()
	cfgFile = "./testdata/test.yaml"

	rootCmd.SetArgs([]string{"-p", ""})

	err := rootCmd.Execute()

	assert.EqualError(t, err, "project name is empty")
}

func TestNewRootCmd_Fails(t *testing.T) {
	setProjectDir()
	cfgFile = "./testdata/test.yaml"
	projectName = "test"

	cmd := NewRootCmd()

	assert.Error(t, cmd.Execute())
}
