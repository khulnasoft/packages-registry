package cmd

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRootCmd(t *testing.T) {
	buff := new(strings.Builder)

	rootCmd.SetOut(buff)
	rootCmd.SetErr(buff)
	rootCmd.SetArgs([]string{})
	err := rootCmd.Execute()

	require.Nil(t, err)

	require.Contains(t, buff.String(), "From this, this tool will configure a KhulnaSoft dynamic child pipeline that will carry out the copy.")
}

func TestRootCmdVersion(t *testing.T) {
	buff := new(strings.Builder)

	rootCmd.SetOut(buff)
	rootCmd.SetErr(buff)
	rootCmd.SetArgs([]string{"--version"})
	err := rootCmd.Execute()

	require.Nil(t, err)

	require.Contains(t, buff.String(), "pkgs_importer version")
	require.Contains(t, buff.String(), "build time")
	require.Contains(t, buff.String(), "commit")
}
