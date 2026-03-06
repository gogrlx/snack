// snack is a unified CLI for system package managers.
package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/fang"
	"github.com/gogrlx/snack"
	"github.com/gogrlx/snack/detect"
	"github.com/spf13/cobra"
)

var (
	version  = "dev"
	flagMgr  string
	flagSudo bool
	flagYes  bool
	flagDry  bool
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "snack",
		Short: "A unified CLI for system package managers",
		Long: `snack wraps system package managers (apt, pacman, apk, dnf, and more)
behind a single, consistent interface.`,
	}

	rootCmd.PersistentFlags().StringVar(&flagMgr, "manager", "", "override auto-detected package manager")
	rootCmd.PersistentFlags().BoolVar(&flagSudo, "sudo", false, "run with sudo")
	rootCmd.PersistentFlags().BoolVar(&flagYes, "yes", false, "assume yes to prompts")
	rootCmd.PersistentFlags().BoolVar(&flagDry, "dry-run", false, "simulate the operation")

	rootCmd.AddCommand(
		installCmd(),
		removeCmd(),
		purgeCmd(),
		upgradeCmd(),
		updateCmd(),
		listCmd(),
		searchCmd(),
		infoCmd(),
		whichCmd(),
		holdCmd(),
		unholdCmd(),
		cleanCmd(),
		repoCmd(),
		keyCmd(),
		groupCmd(),
		detectCmd(),
		versionCmd(),
	)

	if err := fang.Execute(context.Background(), rootCmd); err != nil {
		os.Exit(1)
	}
}

func getManager() (snack.Manager, error) {
	if flagMgr != "" {
		m, err := detect.ByName(flagMgr)
		if err != nil {
			return nil, fmt.Errorf("unknown manager %q", flagMgr)
		}
		return m, nil
	}
	return detect.Default()
}

func opts() []snack.Option {
	var o []snack.Option
	if flagSudo {
		o = append(o, snack.WithSudo())
	}
	if flagYes {
		o = append(o, snack.WithAssumeYes())
	}
	if flagDry {
		o = append(o, snack.WithDryRun())
	}
	return o
}

func targets(args []string, ver string) []snack.Target {
	t := snack.Targets(args...)
	if ver != "" && len(t) > 0 {
		for i := range t {
			t[i].Version = ver
		}
	}
	return t
}

func installCmd() *cobra.Command {
	var ver string
	cmd := &cobra.Command{
		Use:   "install <packages...>",
		Short: "Install packages",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			m, err := getManager()
			if err != nil {
				return err
			}
			_, err = m.Install(cmd.Context(), targets(args, ver), opts()...)
			return err
		},
	}
	cmd.Flags().StringVar(&ver, "version", "", "pin version for all targets")
	return cmd
}

func removeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <packages...>",
		Short: "Remove packages",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			m, err := getManager()
			if err != nil {
				return err
			}
			_, err = m.Remove(cmd.Context(), snack.Targets(args...), opts()...)
			return err
		},
	}
}

func purgeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "purge <packages...>",
		Short: "Purge packages (remove including config files)",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			m, err := getManager()
			if err != nil {
				return err
			}
			return m.Purge(cmd.Context(), snack.Targets(args...), opts()...)
		},
	}
}

func upgradeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade all installed packages",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			m, err := getManager()
			if err != nil {
				return err
			}
			return m.Upgrade(cmd.Context(), opts()...)
		},
	}
}

func updateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "Refresh the package index",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			m, err := getManager()
			if err != nil {
				return err
			}
			return m.Update(cmd.Context())
		},
	}
}

func listCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List installed packages",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			m, err := getManager()
			if err != nil {
				return err
			}
			pkgs, err := m.List(cmd.Context())
			if err != nil {
				return err
			}
			for _, p := range pkgs {
				if p.Version != "" {
					fmt.Printf("%s %s\n", p.Name, p.Version)
				} else {
					fmt.Println(p.Name)
				}
			}
			return nil
		},
	}
}

func searchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "search <query>",
		Short: "Search for packages",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			m, err := getManager()
			if err != nil {
				return err
			}
			pkgs, err := m.Search(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			for _, p := range pkgs {
				line := p.Name
				if p.Version != "" {
					line += " " + p.Version
				}
				if p.Description != "" {
					line += " - " + p.Description
				}
				fmt.Println(line)
			}
			return nil
		},
	}
}

func infoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "info <package>",
		Short: "Show package information",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			m, err := getManager()
			if err != nil {
				return err
			}
			p, err := m.Info(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			fmt.Printf("Name:        %s\n", p.Name)
			fmt.Printf("Version:     %s\n", p.Version)
			if p.Description != "" {
				fmt.Printf("Description: %s\n", p.Description)
			}
			if p.Arch != "" {
				fmt.Printf("Arch:        %s\n", p.Arch)
			}
			if p.Repository != "" {
				fmt.Printf("Repository:  %s\n", p.Repository)
			}
			fmt.Printf("Installed:   %v\n", p.Installed)
			return nil
		},
	}
}

func whichCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "which <path>",
		Short: "Find the package that owns a file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			m, err := getManager()
			if err != nil {
				return err
			}
			fo, ok := m.(snack.FileOwner)
			if !ok {
				return fmt.Errorf("%s does not support file ownership queries", m.Name())
			}
			pkg, err := fo.Owner(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			fmt.Println(pkg)
			return nil
		},
	}
}

func holdCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "hold <packages...>",
		Short: "Hold packages at their current version",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			m, err := getManager()
			if err != nil {
				return err
			}
			h, ok := m.(snack.Holder)
			if !ok {
				return fmt.Errorf("%s does not support hold/unhold", m.Name())
			}
			return h.Hold(cmd.Context(), args)
		},
	}
}

func unholdCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "unhold <packages...>",
		Short: "Remove version hold from packages",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			m, err := getManager()
			if err != nil {
				return err
			}
			h, ok := m.(snack.Holder)
			if !ok {
				return fmt.Errorf("%s does not support hold/unhold", m.Name())
			}
			return h.Unhold(cmd.Context(), args)
		},
	}
}

func cleanCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "clean",
		Short: "Autoremove unused packages and clean cache",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			m, err := getManager()
			if err != nil {
				return err
			}
			c, ok := m.(snack.Cleaner)
			if !ok {
				return fmt.Errorf("%s does not support clean operations", m.Name())
			}
			if err := c.Autoremove(cmd.Context(), opts()...); err != nil {
				return fmt.Errorf("autoremove: %w", err)
			}
			if err := c.Clean(cmd.Context()); err != nil {
				return fmt.Errorf("clean: %w", err)
			}
			return nil
		},
	}
}

func detectCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "detect",
		Short: "Show detected package manager(s) and capabilities",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			all := detect.All()
			if len(all) == 0 {
				fmt.Println("No supported package managers detected.")
				return nil
			}
			def, _ := detect.Default()
			for _, m := range all {
				marker := "  "
				if def != nil && m.Name() == def.Name() {
					marker = "* "
				}
				caps := snack.GetCapabilities(m)
				var capList []string
				if caps.VersionQuery {
					capList = append(capList, "version-query")
				}
				if caps.Hold {
					capList = append(capList, "hold")
				}
				if caps.Clean {
					capList = append(capList, "clean")
				}
				if caps.FileOwnership {
					capList = append(capList, "file-owner")
				}
				if caps.RepoManagement {
					capList = append(capList, "repo")
				}
				if caps.KeyManagement {
					capList = append(capList, "keys")
				}
				if caps.Groups {
					capList = append(capList, "groups")
				}
				if caps.NameNormalize {
					capList = append(capList, "normalize")
				}
				if caps.PackageUpgrade {
					capList = append(capList, "pkg-upgrade")
				}
				capStr := ""
				if len(capList) > 0 {
					capStr = " [" + strings.Join(capList, ", ") + "]"
				}
				fmt.Printf("%s%s%s\n", marker, m.Name(), capStr)
			}
			fmt.Println("\n* = default")
			return nil
		},
	}
}

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show snack version",
		Args:  cobra.NoArgs,
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Printf("snack %s\n", version)
		},
	}
}

func repoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "repo",
		Short: "Manage package repositories",
	}
	cmd.AddCommand(repoListCmd(), repoAddCmd(), repoRemoveCmd())
	return cmd
}

func repoListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List configured repositories",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			m, err := getManager()
			if err != nil {
				return err
			}
			rm, ok := m.(snack.RepoManager)
			if !ok {
				return fmt.Errorf("%s does not support repository management", m.Name())
			}
			repos, err := rm.ListRepos(cmd.Context())
			if err != nil {
				return err
			}
			for _, r := range repos {
				status := "disabled"
				if r.Enabled {
					status = "enabled"
				}
				fmt.Printf("%s %s [%s]\n", r.ID, r.URL, status)
			}
			return nil
		},
	}
}

func repoAddCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add <url>",
		Short: "Add a package repository",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			m, err := getManager()
			if err != nil {
				return err
			}
			rm, ok := m.(snack.RepoManager)
			if !ok {
				return fmt.Errorf("%s does not support repository management", m.Name())
			}
			repo := snack.Repository{
				URL:     args[0],
				Enabled: true,
			}
			return rm.AddRepo(cmd.Context(), repo)
		},
	}
}

func repoRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <id>",
		Short: "Remove a package repository",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			m, err := getManager()
			if err != nil {
				return err
			}
			rm, ok := m.(snack.RepoManager)
			if !ok {
				return fmt.Errorf("%s does not support repository management", m.Name())
			}
			return rm.RemoveRepo(cmd.Context(), args[0])
		},
	}
}

func keyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "key",
		Short: "Manage GPG signing keys",
	}
	cmd.AddCommand(keyListCmd(), keyAddCmd(), keyRemoveCmd())
	return cmd
}

func keyListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List trusted signing keys",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			m, err := getManager()
			if err != nil {
				return err
			}
			km, ok := m.(snack.KeyManager)
			if !ok {
				return fmt.Errorf("%s does not support key management", m.Name())
			}
			keys, err := km.ListKeys(cmd.Context())
			if err != nil {
				return err
			}
			for _, k := range keys {
				fmt.Println(k)
			}
			return nil
		},
	}
}

func keyAddCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add <key>",
		Short: "Add a GPG signing key (URL, file path, or key ID)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			m, err := getManager()
			if err != nil {
				return err
			}
			km, ok := m.(snack.KeyManager)
			if !ok {
				return fmt.Errorf("%s does not support key management", m.Name())
			}
			return km.AddKey(cmd.Context(), args[0])
		},
	}
}

func keyRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <key-id>",
		Short: "Remove a GPG signing key",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			m, err := getManager()
			if err != nil {
				return err
			}
			km, ok := m.(snack.KeyManager)
			if !ok {
				return fmt.Errorf("%s does not support key management", m.Name())
			}
			return km.RemoveKey(cmd.Context(), args[0])
		},
	}
}

func groupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "group",
		Short: "Manage package groups",
	}
	cmd.AddCommand(groupListCmd(), groupInfoCmd(), groupInstallCmd())
	return cmd
}

func groupListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List available package groups",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			m, err := getManager()
			if err != nil {
				return err
			}
			g, ok := m.(snack.Grouper)
			if !ok {
				return fmt.Errorf("%s does not support package groups", m.Name())
			}
			groups, err := g.GroupList(cmd.Context())
			if err != nil {
				return err
			}
			for _, grp := range groups {
				fmt.Println(grp)
			}
			return nil
		},
	}
}

func groupInfoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "info <group>",
		Short: "Show packages in a group",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			m, err := getManager()
			if err != nil {
				return err
			}
			g, ok := m.(snack.Grouper)
			if !ok {
				return fmt.Errorf("%s does not support package groups", m.Name())
			}
			pkgs, err := g.GroupInfo(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			for _, p := range pkgs {
				fmt.Println(p.Name)
			}
			return nil
		},
	}
}

func groupInstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "install <group>",
		Short: "Install all packages in a group",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			m, err := getManager()
			if err != nil {
				return err
			}
			g, ok := m.(snack.Grouper)
			if !ok {
				return fmt.Errorf("%s does not support package groups", m.Name())
			}
			return g.GroupInstall(cmd.Context(), args[0], opts()...)
		},
	}
}
