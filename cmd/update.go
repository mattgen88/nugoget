package cmd

import (
	"fmt"
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
)

var (
	projectRegexp         = regexp.MustCompile(`^(?:The given )?[Pp]roject ` + "`" + `([\w.]+)` + "`")
	versionRegexp         = regexp.MustCompile(`^\s+\[([\w\d.]+)\]:`)
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
	// execute and capture nuget output
	//dotnet list package --outdated
	// default to appending --highest-minor
	// flag to override --highest-minor to --highest-patch or remove flag entirely
	// parse output

	//	The following sources were used:
	//	https://nexus.infra.acvauctions.com/repository/nuget-group/
	//	https://api.nuget.org/v3/index.json
	//
	//  Project `Documents.Service` has the following updates to its packages
	// 	[netcoreapp3.1]:
	// 	Top-level Package      Requested   Resolved   Latest
	// 	> NewRelic.Agent       8.35.0      8.35.0     8.36.0
	//
	//  The given project `Documents.Domain` has no updates given the current sources.
	//  The given project `Documents.Dtos` has no updates given the current sources.
	//  Project `Documents.Infrastructure` has the following updates to its packages
	// 	[netcoreapp3.1]:
	// 	Top-level Package               Requested   Resolved   Latest
	// 	> ACV.Hydra.Shared.Workers      4.9.1       4.9.1      4.9.2
	// 	> AWSSDK.Core                   3.5.1.42    3.5.1.42   3.5.1.48
	// 	> AWSSDK.S3                     3.5.5.2     3.5.5.2    3.5.6.4
	//
	//  Project `Documents.Tests` has the following updates to its packages
	// 	[netcoreapp3.1]:
	// 	Top-level Package                      Requested   Resolved   Latest
	// 	> ACV.Hydra.Shared.Testing.Common      3.6.11      3.6.11     3.6.12
	// 	> Microsoft.NET.Test.Sdk               16.8.0      16.8.0     16.8.3
	//
	//  Project `Documents.Application` has the following updates to its packages
	// 	[netcoreapp3.1]:
	// 	Top-level Package                               Requested   Resolved   Latest
	// 	> ACV.Hydra.DataHub.BusinessDocumentEvents      1.1.0       1.1.0      1.1.1
	//
	//  Project `Documents.EventListener.Worker` has the following updates to its packages
	// 	[netcoreapp3.1]:
	// 	Top-level Package                           Requested   Resolved   Latest
	// 	> ACV.Hydra.DataHub.OrganizationEvents      1.3.8       1.3.8      1.3.13
	// 	> ACV.Hydra.Shared.Workers                  4.9.1       4.9.1      4.9.2
	// 	> NewRelic.Agent                            8.35.0      8.35.0     8.36.0
	//
	//  Project `Documents.DealerDocs.Worker` has the following updates to its packages
	// 	[netcoreapp3.1]:
	// 	Top-level Package                               Requested   Resolved   Latest
	// 	> ACV.Hydra.DataHub.BusinessDocumentEvents      1.1.0       1.1.0      1.1.1
	// 	> ACV.Hydra.Shared.Workers                      4.9.1       4.9.1      4.9.2
	// 	> NewRelic.Agent                                8.35.0      8.35.0     8.36.0
	//
	//  The given project `Documents.Configuration` has no updates given the current sources.
	//  The given project `Documents.DatabaseMigrator` has no updates given the current sources.
	//  Project `Documents.Tests.API` has the following updates to its packages
	// 	[netcoreapp3.1]:
	// 	Top-level Package                      Requested   Resolved   Latest
	// 	> ACV.Hydra.Shared.Testing.Common      3.6.11      3.6.11     3.6.12
	// 	> Microsoft.NET.Test.Sdk               16.8.0      16.8.0     16.8.3
	//

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
		nugetArgs = append(nugetArgs, "--highest-minor")
	}

	nugetCmd := exec.Command(path, nugetArgs...)

	out, err := nugetCmd.CombinedOutput()
	if err != nil {
		fmt.Println(err)
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
			// fmt.Println("Project line")
			break
		case versionRegexp.MatchString(line):
			// fmt.Println("Version line")
			matches := versionRegexp.FindStringSubmatch(line)
			currentProject.version = newVersion()
			currentProject.version.name = matches[1]
			break
		case headerRegexp.MatchString(line):
			// fmt.Println("Header line")
			break
		case packageVersionsRegexp.MatchString(line):
			// fmt.Println("Package Versions String")
			matches := packageVersionsRegexp.FindStringSubmatch(line)
			currentProject.version.packageData[matches[1]] = newPackageData(matches[1], matches[2], matches[3], matches[4])
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
					panic(err)
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
}
