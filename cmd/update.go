package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

var (
	updateCmd = &cobra.Command{
		Use:   "update",
		Short: "Updates the dependencies",
		Long:  `For each project found, updates the dependencies`,
		Run:   update,
	}
	major  = false
	patch  = false
	dryRun = false
	locks  []string
)

const (
	ExitBadLock = iota
	ExitBadCallOut
)

var (
	projectRegexp         = regexp.MustCompile(`^(?:The given )?[Pp]roject ` + "`" + `([\w.]+)` + "`")
	versionRegexp         = regexp.MustCompile(`^\s+\[([\w\d.]+)]:`)
	headerRegexp          = regexp.MustCompile(`^\s+Top-level Package\s+Requested\s+Resolved\s+Latest`)
	packageVersionsRegexp = regexp.MustCompile(`^\s+>\s+([\w.-]+)\s+([\d\w.-]+)\s+([\d\w.-]+)\s+([\d\w.-]+)`)
	blankRegexp           = regexp.MustCompile(`$^`)
)

type project struct {
	name    string
	version *version
}

type version struct {
	name        string
	packageData map[string]*packageData
}

type packageData struct {
	name      string
	requested string
	resolved  string
	latest    string
}

func newProject(name string) *project {
	p := &project{
		name: name,
	}
	p.version = newVersion()
	return p
}

func newVersion() *version {
	v := &version{}
	v.packageData = make(map[string]*packageData)
	return v
}

func newPackageData(name, requested, resolved, latest string) *packageData {
	p := &packageData{
		name:      name,
		requested: requested,
		resolved:  resolved,
		latest:    latest,
	}
	return p
}

func update(cmd *cobra.Command, args []string) {
	var lockedPackages = make(map[string]string)
	if len(locks) > 0 {
		// each lock should be in the pattern of <name>#<version>
		for _, lock := range locks {
			if !strings.Contains(lock, "#") {
				fmt.Printf("Invalid lock %s should be in format <name>#<version>\n", lock)
				os.Exit(ExitBadLock)
			}
			parts := strings.Split(lock, "#")
			pkg := parts[0]
			version := parts[1]
			lockedPackages[pkg] = version
		}
	}
	if major {
		fmt.Println("Will do major upgrades!")
	}
	if dryRun {
		fmt.Println("No changes will be made, dry run enabled")
	}

	path, err := exec.LookPath("dotnet")
	if err != nil {
		panic("Try installing dotnet first and making sure it is in your PATH")
	}
	nugetArgs := []string{"list", "package", "--outdated"}
	if major {
		// append nothing
	} else if patch {
		nugetArgs = append(nugetArgs, "--highest-patch")
	} else {
		nugetArgs = append(nugetArgs, "--highest-minor")
	}

	if debug {
		fmt.Println("Executing: ", path, nugetArgs)
	}
	nugetCmd := exec.Command(path, nugetArgs...)

	out, err := nugetCmd.CombinedOutput()
	if err != nil {
		fmt.Println(err)
		os.Exit(ExitBadCallOut)
		panic(err)
	}

	lines := strings.Split(string(out), "\n")

	var projects map[string]*project
	projects = make(map[string]*project)
	var currentProject *project
	updates := 0

	for _, line := range lines {
		switch true {
		case projectRegexp.MatchString(line):
			matches := projectRegexp.FindStringSubmatch(line)
			p := newProject(matches[1])
			projects[p.name] = p
			currentProject = p
			fmt.Printf("Checking for updates in %s\n", p.name)
			break
		case versionRegexp.MatchString(line):
			matches := versionRegexp.FindStringSubmatch(line)
			currentProject.version = newVersion()
			currentProject.version.name = matches[1]
			break
		case headerRegexp.MatchString(line):
			break
		case packageVersionsRegexp.MatchString(line):
			matches := packageVersionsRegexp.FindStringSubmatch(line)
			if _, ok := lockedPackages[matches[1]]; ok {
				currentProject.version.packageData[matches[1]] = newPackageData(matches[1], lockedPackages[matches[1]], lockedPackages[matches[1]], lockedPackages[matches[1]])
				updates++
			} else {
				currentProject.version.packageData[matches[1]] = newPackageData(matches[1], matches[2], matches[3], matches[4])
			}
			if matches[2] != matches[4] {
				updates++
			}
			break
		case blankRegexp.MatchString(line):
			break
		default:

		}
	}

	fmt.Printf("%d updates found\n", updates)

	for _, project := range projects {
		fmt.Printf("\n\n%s\n", project.name)
		for _, dependency := range project.version.packageData {
			if dependency.requested != dependency.latest {
				fmt.Printf("%s will be updated from %s to %s\n", dependency.name, dependency.requested, dependency.latest)
				if dryRun {
					continue
				}
				nugetCmd = exec.Command(path, "add", project.name, "package", dependency.name, "-v", dependency.latest)
				out, err = nugetCmd.CombinedOutput()
				if err != nil {
					os.Exit(ExitBadCallOut)
				}

				fmt.Println(string(out))
			}
		}
	}

}

func init() {
	updateCmd.PersistentFlags().BoolVarP(&major, "major", "M", false, "When true, will update to latest major. This can break your project!")
	updateCmd.PersistentFlags().BoolVarP(&patch, "patch", "P", false, "When true, will update to latest patch.")
	updateCmd.PersistentFlags().BoolVarP(&dryRun, "dryrun", "D", false, "When true, will only do a dry run, no changes will be made.")
	updateCmd.PersistentFlags().StringArrayVarP(&locks, "lock", "l", []string{}, "list of locked packages and their versions in the format of <packagename>#version. All instances of each package will be set to that version specifically")
}
