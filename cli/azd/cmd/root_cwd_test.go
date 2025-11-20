// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License.

package cmd

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/azure/azure-dev/cli/azd/internal"
	"github.com/azure/azure-dev/cli/azd/pkg/ioc"
	"github.com/stretchr/testify/require"
)

func Test_CwdDirectoryCreation_WithNoPrompt(t *testing.T) {
	tempDir := t.TempDir()
	newDir := filepath.Join(tempDir, "new-directory")

	// Ensure the directory doesn't exist
	_, err := os.Stat(newDir)
	require.True(t, os.IsNotExist(err), "Directory should not exist before test")

	// Create root command with NoPrompt=true
	rootContainer := ioc.NewNestedContainer(nil)
	ctx := context.Background()
	ioc.RegisterInstance(rootContainer, ctx)

	opts := &internal.GlobalCommandOptions{
		Cwd:      newDir,
		NoPrompt: true,
	}
	ioc.RegisterInstance(rootContainer, opts)

	rootCmd := NewRootCmd(false, nil, rootContainer)

	// Execute the PersistentPreRunE
	err = rootCmd.PersistentPreRunE(rootCmd, []string{})
	require.NoError(t, err, "PersistentPreRunE should succeed")

	// Verify the directory was created
	stat, err := os.Stat(newDir)
	require.NoError(t, err, "Directory should exist after PersistentPreRunE")
	require.True(t, stat.IsDir(), "Created path should be a directory")

	// Verify we changed to the new directory
	currentDir, err := os.Getwd()
	require.NoError(t, err)
	require.Equal(t, newDir, currentDir, "Current directory should be the new directory")

	// Clean up by running PersistentPostRunE to restore original directory
	err = rootCmd.PersistentPostRunE(rootCmd, []string{})
	require.NoError(t, err)
}

func Test_CwdDirectoryExists(t *testing.T) {
	tempDir := t.TempDir()
	existingDir := filepath.Join(tempDir, "existing-directory")

	// Create the directory first
	err := os.Mkdir(existingDir, 0755)
	require.NoError(t, err)

	// Create root command
	rootContainer := ioc.NewNestedContainer(nil)
	ctx := context.Background()
	ioc.RegisterInstance(rootContainer, ctx)

	opts := &internal.GlobalCommandOptions{
		Cwd:      existingDir,
		NoPrompt: false,
	}
	ioc.RegisterInstance(rootContainer, opts)

	rootCmd := NewRootCmd(false, nil, rootContainer)

	// Execute the PersistentPreRunE
	err = rootCmd.PersistentPreRunE(rootCmd, []string{})
	require.NoError(t, err, "PersistentPreRunE should succeed for existing directory")

	// Verify we changed to the existing directory
	currentDir, err := os.Getwd()
	require.NoError(t, err)
	require.Equal(t, existingDir, currentDir, "Current directory should be the existing directory")

	// Clean up by running PersistentPostRunE to restore original directory
	err = rootCmd.PersistentPostRunE(rootCmd, []string{})
	require.NoError(t, err)
}

func Test_CwdNotSet(t *testing.T) {
	// Create root command without setting Cwd
	rootContainer := ioc.NewNestedContainer(nil)
	ctx := context.Background()
	ioc.RegisterInstance(rootContainer, ctx)

	opts := &internal.GlobalCommandOptions{
		Cwd:      "",
		NoPrompt: false,
	}
	ioc.RegisterInstance(rootContainer, opts)

	rootCmd := NewRootCmd(false, nil, rootContainer)

	// Get current directory before executing
	beforeDir, err := os.Getwd()
	require.NoError(t, err)

	// Execute the PersistentPreRunE
	err = rootCmd.PersistentPreRunE(rootCmd, []string{})
	require.NoError(t, err, "PersistentPreRunE should succeed when Cwd is not set")

	// Verify we're still in the same directory
	afterDir, err := os.Getwd()
	require.NoError(t, err)
	require.Equal(t, beforeDir, afterDir, "Directory should not change when Cwd is not set")

	// Clean up by running PersistentPostRunE (should be no-op)
	err = rootCmd.PersistentPostRunE(rootCmd, []string{})
	require.NoError(t, err)
}
